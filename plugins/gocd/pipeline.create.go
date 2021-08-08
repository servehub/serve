package gocd

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"sort"

	"github.com/servehub/serve/manifest"
	"github.com/servehub/utils/gabs"
)

func init() {
	manifest.PluginRegestry.Add("gocd.pipeline.create", goCdPipelineCreate{})
}

/**
 * plugin for manifest section "goCd.pipeline.create"
 * section structure:
 *
 * goCd.pipeline.create:
 *   api-url: goCd_URL
 *   environment: ENV
 *   branch: BRANCH
 *   allowed-branches: [BRANCH, ...]
 *   pipeline:
 *     group: GROUP
 *     pipeline:
 *       according to the description: https://api.go.cd/current/#the-pipeline-config-object
 */

type goCdCredents struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type goCdPipelineCreate struct{}

func (p goCdPipelineCreate) Run(data manifest.Manifest) error {
	if suffix := data.GetString("name-suffix"); suffix != "" {
		data.Set("pipeline.pipeline.name", data.GetString("pipeline.pipeline.name")+suffix)
		data.Set("pipeline.pipeline.envs.SERVE_EXTRA_ARGS.value", data.GetStringOr("pipeline.pipeline.envs.SERVE_EXTRA_ARGS.value", "")+" --var name-suffix="+suffix)
	}

	name := data.GetString("pipeline.pipeline.name")
	url := data.GetString("api-url")
	if data.GetString("pipeline.pipeline.template") == "" {
		data.DelTree("pipeline.pipeline.template")
	}

	replaceMapWithArray(data, "pipeline.pipeline.envs", "pipeline.pipeline.environment_variables")
	replaceMapWithArray(data, "pipeline.pipeline.params", "pipeline.pipeline.parameters")

	depends := []string{}
	if !data.Has("pipeline.pipeline.materials") {
		data.Set("pipeline.pipeline.materials", []interface{}{})
	}
	for _, dep := range data.GetArray("depends") {
		pipeline := dep.GetString("pipeline")
		depends = append(depends, pipeline)
		data.ArrayAppend("pipeline.pipeline.materials",
			map[string]interface{}{"type": "dependency",
				"attributes": map[string]interface{}{
					"name":        dep.GetStringOr("name", pipeline),
					"pipeline":    pipeline,
					"stage":       data.GetStringOr("stage", "Build"),
					"auto_update": true}})
	}

	branch := data.GetString("branch")

	m := false
	for _, b := range data.GetArray("allowed-branches") {
		re := b.Unwrap().(string)
		if re == "*" || re == branch {
			m = true
			break
		} else if m, _ = regexp.MatchString(re, branch); m {
			break
		}
	}

	if !m {
		log.Println("branch ", branch, " not in ", data.GetString("allowed-branches"))
		return nil
	}

	if data.GetBool("purge") {
		return goCdDelete(name, data.GetString("environment"), url,
			map[string]string{"Accept": "application/vnd.go.cd.v11+json"})
	}

	exist, err := goCdRequest("GET", url+"/go/api/admin/pipelines/"+name, "",
		map[string]string{"Accept": "application/vnd.go.cd.v11+json"})
	if err != nil {
		return err
	}

	if exist.StatusCode == http.StatusOK {
		pipeline := data.GetTree("pipeline.pipeline")
		pipeline.Set("group", data.GetString("pipeline.group"))
		err = goCdUpdate(name, data.GetString("environment"), url, pipeline.String(),
			map[string]string{"If-Match": exist.Header.Get("ETag"), "Accept": "application/vnd.go.cd.v11+json"}, depends)

	} else if exist.StatusCode == http.StatusNotFound {
		err = goCdCreate(name, data.GetString("environment"), url, data.GetTree("pipeline").String(),
			map[string]string{"Accept": "application/vnd.go.cd.v11+json"})

	} else {
		return fmt.Errorf("Operation error: %s", exist.Status)
	}

	if err != nil {
		return err
	}

	return nil
}

func goCdCreate(name string, env string, resource string, body string, headers map[string]string) error {
	if resp, err := goCdRequest("POST", resource+"/go/api/admin/pipelines", body, headers); err != nil {
		return err
	} else if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()
		if body, err := ioutil.ReadAll(resp.Body); err == nil {
			log.Printf("Operation body: %s", body)
		}
		return fmt.Errorf("Operation error: %s", resp.Status)
	}

	return gocdChangeEnv(resource, env, map[string]interface{}{"add": []string{name}})
}

func goCdUpdate(name string, env string, resource string, body string, headers map[string]string, depends []string) error {
	if resp, err := goCdRequest("PUT", resource+"/go/api/admin/pipelines/"+name, body, headers); err != nil {
		return err
	} else if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()
		if body, err := ioutil.ReadAll(resp.Body); err == nil {
			log.Printf("Operation body: %s", body)
		}
		return fmt.Errorf("Operation error: %s", resp.Status)
	}

	if currentEnv, err := goCdFindEnv(resource, name, depends); err == nil {
		if env != currentEnv {
			if currentEnv != "" {
				if err := gocdChangeEnv(resource, currentEnv, map[string]interface{}{"remove": []string{name}}); err != nil {
					return err
				}
			}

			if err := gocdChangeEnv(resource, env, map[string]interface{}{"add": []string{name}}); err != nil {
				return err
			}
		}
	} else {
		return err
	}

	return nil
}

func goCdSchedule(name string, resource string, body string, headers map[string]string) error {
	_, err := goCdRequest("POST", resource + "/go/api/pipelines/" + name + "/schedule", body, headers)
	return err
}

func goCdDelete(name string, env string, resource string, headers map[string]string) error {
	if err := gocdChangeEnv(resource, env, map[string]interface{}{"remove": []string{name}}); err != nil {
		log.Println("Error on remove pipeline from env: ", err)
	}

	if resp, err := goCdRequest("DELETE", resource+"/go/api/admin/pipelines/"+name, "", headers); err != nil {
		return err
	} else if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Operation error: %s", resp.Status)
	}

	return nil
}

func gocdChangeEnv(apiUrl string, env string, actions map[string]interface{}) error {
	body, _ := json.Marshal(map[string]interface{}{
		"pipelines": actions,
	})

	resp, err := goCdRequest("PATCH", apiUrl+"/go/api/admin/environments/"+env, string(body),
		map[string]string{"Accept": "application/vnd.go.cd.v3+json"})
	if err != nil {
		return err
	} else if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Operation error: %s", resp.Status)
	}

	return nil
}

func goCdFindEnv(resource string, pipeline string, depends []string) (string, error) {
	resp, err := goCdRequest("GET", resource+"/go/api/admin/environments", "",
		map[string]string{"Accept": "application/vnd.go.cd.v3+json"})
	if err != nil {
		return "", err
	} else if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Operation error: %s", resp.Status)
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	tree, err := gabs.ParseJSON(body)
	if err != nil {
		return "", err
	}

	sort.Strings(depends)
	envs, _ := tree.Path("_embedded.environments").Children()
	curEnvName := ""
	for _, env := range envs {
		envName := env.Path("name").Data().(string)
		pipelines, _ := env.Path("pipelines").Children()

		if len(depends) > 0 {
			if i := sort.SearchStrings(depends, envName); i != len(depends) {
				depends = append(depends[:i], depends[i+1:]...)
			}
		} else {
			if curEnvName != "" {
				break
			}
		}

		for _, pline := range pipelines {
			if pline.Path("name").Data().(string) == pipeline {
				curEnvName = envName
			}
		}
	}

	if len(depends) != 0 {
		return curEnvName, fmt.Errorf("not found depends: %v", depends)
	}

	return curEnvName, nil
}

var httpClient = &http.Client{Transport: &http.Transport{
	TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
}}

var goCdRequest = func(method string, resource string, body string, headers map[string]string) (*http.Response, error) {
	req, _ := http.NewRequest(method, resource, bytes.NewReader([]byte(body)))

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	req.Header.Set("Content-Type", "application/json")

	data, err := ioutil.ReadFile("/etc/serve/gocd_credentials")
	if err != nil {
		return nil, fmt.Errorf("Credentias file error: %v", err)
	}

	creds := &goCdCredents{}
	json.Unmarshal(data, creds)

	req.SetBasicAuth(creds.Login, creds.Password)

	log.Printf(" --> %s %s:\n%s\n%s\n\n", method, resource, req.Header, body)

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	log.Printf("<-- %s\n", resp.Status)
	return resp, nil
}

func replaceMapWithArray(data manifest.Manifest, mapPath string, arrPath string) {
	arrs := make([]interface{}, 0)
	for k, v := range data.GetMap(mapPath) {
		v.Set("name", k)
		arrs = append(arrs, v.Unwrap())
	}
	data.Set(arrPath, arrs)
	data.DelTree(mapPath)
}
