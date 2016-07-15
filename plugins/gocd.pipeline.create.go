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
)

func init() {
	manifest.PluginRegestry.Add("gocd.pipeline.create", GoCdPipelineCreate{})
}

/**
 * plugin for manifest section "gocd.pipeline.create"
 * section structure:
 *
 * gocd.pipeline.create:
 * 	login: LOGIN
 * 	password: PASSWORD
 * 	url: GOCD_URL
 *  pipeline_name: NAME
 * 	pipeline:
 * 		group: GROUP
 * 		pipeline:
 * 			according to the description: https://api.go.cd/current/#the-pipeline-config-object
 */
type GoCdPipelineCreate struct{}

func (p GoCdPipelineCreate) Run(data manifest.Manifest) error {
	url := data.GetString("url") + "/" + data.GetString("pipeline_name")
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

	resp, err := gocdRequest("GET", url, "", nil)
	if err != nil {
		log.Println(err)
		return err
	}

	if resp.StatusCode == http.StatusOK {
		resp, err = gocdRequest("PUT", url, body, map[string]string{"If-Match": resp.Header.Get("ETag")})
	} else if resp.StatusCode == http.StatusNotFound {
		resp, err = gocdRequest("POST", data.GetString("url"), body, nil)
	} else {
		log.Println("Operation error: " + resp.Status)
		return errors.New("Operation error: " + resp.Status)
	}

	if err != nil {
		log.Println(err)
		return err
	}

	if resp.StatusCode != http.StatusOK {
		log.Println("Operation error: " + resp.Status)
		return errors.New("Operation error: " + resp.Status)
	}

	return nil
}

func gocdRequest(method string, resource string, body string, headers map[string]string) (*http.Response, error) {
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
