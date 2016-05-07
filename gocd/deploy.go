package gocd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"

	"github.com/kulikov/serve/manifest"
)

type (
	DeployPlugin struct{}

	DeployManifest struct {
		Build  []DeployBuild `yaml:"build"`
		Deploy Deploy        `yaml:"deploy"`
	}

	DeployBuild struct {
		Debian *BuildDebian `yaml:"debian"`
	}

	BuildDebian struct{}

	Deploy struct {
		Debian *DeployDebian `yaml:"debian"`
	}

	DeployDebian struct {
		Cluster map[string]string `yaml:"cluster"`
	}
)

func (p Pipeline) addStage(s Stage) {
	p.Stages = append(p.Stages, s)
}

func (ea DeployPlugin) Run(conf *viper.Viper, manf *manifest.Manifest) error {
	depl := DeployManifest{}
	yaml.Unmarshal(manf.Source, &depl)

	pipeline := Pipeline{
		Name: manf.Info.Name,
		Materials: []Material{
			Material{
				Type: "git",
				Attributes: MaterialAttrs{
					Url:         manf.GitSshUrl,
					Destination: "sources",
				},
			},
		},
	}

	for _, build := range depl.Build {
		if build.Debian != nil {
			pipeline.addStage(Stage{
				Name: "Build",
				CleanWorkingDirectory: true,
				FetchMaterials:        true,
				Jobs: []Job{
					Job{
						Name:      "Create-Package",
						Resources: []string{"Builder", "Debian"},
						Tasks: []Task{execTask(
							"/var/go/inn-ci-tools/build-package.sh --suffix=-master --patch-version=$GO_PIPELINE_LABEL --repo=" + manf.GitSshUrl + " --distribution=unstable",
						)},
					},
				},
			})
		}
	}

	if depl.Deploy.Debian != nil {
		deb := depl.Deploy.Debian
		project := "inn-" + manf.Info.Name
		qaNodes := deb.Cluster["qa"]
		liveNodes := deb.Cluster["live"]

		if qaNodes != "" {
			pipeline.addStage(Stage{
				Name: "Deployment-QA",
				Jobs: []Job{
					Job{
						Name:      "Deploy",
						Resources: []string{"Builder", "Debian"},
						Tasks: []Task{execTask(
							"dig +short "+qaNodes+" | sort | uniq | parallel -j 10",
							"/var/go/inn-ci-tools/go/go-package-deploy.sh --target={} --project="+project+" --version="+manf.Info.Version+".$GO_PIPELINE_LABEL --purge-pattern="+project+"-v.*",
						)},
					},
				},
			})

			pipeline.addStage(Stage{
				Name: "Ready-to-Ship",
				Approval: Approval{
					Type: "manual",
				},
				Jobs: []Job{
					Job{
						Name:      "Approve-Package",
						Resources: []string{"Builder", "Debian"},
						Tasks: []Task{execTask(
							"/var/go/inn-ci-tools/go/go-package-approve.sh",
							"--suffix=-master --patch-version=$GO_PIPELINE_LABEL --project="+project+" --version="+manf.Info.Version+" --src-repo=unstable --dst-repo=stable",
						)},
					},
				},
			})
		}

		if liveNodes != "" {
			pipeline.addStage(Stage{
				Name: "Deployment-Live",
				Approval: Approval{
					Type: "manual",
					Authorization: Authorization{
						Roles: []string{"QA"},
					},
				},
				Jobs: []Job{
					Job{
						Name:      "Deploy",
						Resources: []string{"Builder", "Debian"},
						Tasks: []Task{execTask(
							"dig +short "+liveNodes+" | sort | uniq | parallel -j 10",
							"/var/go/inn-ci-tools/go/go-package-deploy.sh",
							"--target={} --project="+project+" --version="+manf.Info.Version+".$GO_PIPELINE_LABEL --purge-pattern="+project+"-v.*",
						)},
					},
				},
			})
		}
	}

	resp, err := requestGo("GET", "/admin/pipelines/"+manf.Info.Name, nil, nil)
	if err != nil {
		return err
	}

	if resp.StatusCode == 200 {
		body, _ := json.Marshal(pipeline)
		_, err := requestGo("PUT", "/admin/pipelines/"+manf.Info.Name, body, map[string]string{"If-Match": resp.Header.Get("ETag")})
		return err
	} else if resp.StatusCode == 404 {
		body, _ := json.Marshal(CreatePipline{"other", pipeline})
		_, err := requestGo("POST", "/admin/pipelines", body, nil)
		return err
	} else {
		return fmt.Errorf("Error %v", resp.Status)
	}
}

func requestGo(method string, resource string, body []byte, headers map[string]string) (*http.Response, error) {
	req, _ := http.NewRequest(method, "https://go.inn.ru/go/api"+resource, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/vnd.go.cd.v1+json")
	req.SetBasicAuth("login", "pass")

	return http.DefaultClient.Do(req)
}

func execTask(cmd ...string) Task {
	return Task{
		Type: "exec",
		Attributes: TaskAttributes{
			RunIf:            []string{"passed"},
			WorkingDirectory: "sources",
			Command:          "/bin/bash",
			Arguments:        []string{"-c", strings.Join(cmd, " ")},
		},
	}
}
