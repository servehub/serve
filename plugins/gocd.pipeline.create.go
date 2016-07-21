package plugins

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"github.com/InnovaCo/serve/manifest"
	"regexp"
	"log"
	//"net/url"
)

func init() {
	manifest.PluginRegestry.Add("gocd.pipeline.create", goCdPipelineCreate{})
}

/**
 * plugin for manifest section "goCd.pipeline.create"
 * section structure:
 *
 * goCd.pipeline.create:
 * 	login: LOGIN
 * 	password: PASSWORD
 * 	url: goCd_URL
 *  pipeline_name: NAME
 *  environment: ENV
 *  allowed_branches: [BRANCH, ...]
 * 	pipeline:
 * 		group: GROUP
 * 		pipeline:
 * 			according to the description: https://api.go.cd/current/#the-pipeline-config-object
 */
type goCdPipelineCreate struct{}

func (p goCdPipelineCreate) Run(data manifest.Manifest) error {
	name := data.GetString("pipeline_name")
	url := data.GetString("url")
	body := data.GetTree("pipeline").String()
	branch := data.GetString("branch")

	m := false
	for _, b := range data.GetArray("allowed_branches") {
		re := b.Unwrap().(string)
		if re == branch {
			m = true
			break
		} else if m, _ = regexp.MatchString(re, branch); m {
			break
		}
	}

	if !m {
		log.Println("branch ", branch, " not in ", data.GetString("allowed_branches"))
		return errors.New("branch " + branch + " not in " + data.GetString("allowed_branches"))
	}

	resp, err := goCdRequest("GET", url + "/pipelines/" + name, "", nil)
	if err != nil {
		log.Println(err)
		return err
	}

	if resp.StatusCode == http.StatusOK {
		err = goCdUpdate(name, data.GetString("environment"), url, body, map[string]string{"If-Match": resp.Header.Get("ETag")})
	} else if resp.StatusCode == http.StatusNotFound {
		err = goCdCreate(name, data.GetString("environment"), url, body, nil)
	} else {
		log.Println("Operation error: " + resp.Status)
		return errors.New("Operation error: " + resp.Status)
	}

	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

func goCdCreate(name string, env string, resource string, body string, headers map[string]string) error {
	if resp, err := goCdRequest("POST", resource + "/pipelines", body, nil); err != nil {
		return err
	} else if resp.StatusCode != http.StatusOK {
		return errors.New("Operation error: " + resp.Status)
	}

	//val := url.Values{}
	//val.Add("pipelines", name)
	//
	//if resp, err := goCdRequest("PATH", resource + "/environments/" + env , val.Encode(), headers); err != nil {
	//	return err
	//} else if resp.StatusCode != http.StatusOK {
	//	return errors.New("Operation error: " + resp.Status)
	//}

	return nil
}

func goCdUpdate(name string, env string, resource string, body string, headers map[string]string) error {
	if resp, err := goCdRequest("PUT", resource + "/pipelines/" + name , body, headers); err != nil {
		return err
	} else if resp.StatusCode != http.StatusOK {
		return errors.New("Operation error: " + resp.Status)
	}

	return nil
}

func goCdDelete(name string, env string, resource string, body string, headers map[string]string) error {
	if resp, err := goCdRequest("DELETE", resource + "/pipeline/" + name, "", nil); err != nil {
		return err
	} else if resp.StatusCode != http.StatusOK {
		return errors.New("Operation error: " + resp.Status)
	}

	return nil
}

func goCdRequest(method string, resource string, body string, headers map[string]string) (*http.Response, error) {
	req, _ := http.NewRequest(method, resource, bytes.NewReader([]byte(body)))

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	req.Header.Set("Accept", "application/vnd.go.cd.v1+json")
	req.Header.Set("Content-Type", "application/json")

	data, err := ioutil.ReadFile("/etc/serve/gocd_credentials")
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Credentias file error: %v", err))
	}

	creds := &goCdCredents{}
	json.Unmarshal(data, creds)

	req.SetBasicAuth(creds.Login, creds.Password)

	log.Printf(" --> %s %s:\n%s\n", method, resource, body)

	return http.DefaultClient.Do(req)
}

type goCdCredents struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}
