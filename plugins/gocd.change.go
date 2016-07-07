package plugins

import (
	"github.com/InnovaCo/serve/manifest"
	"net/http"
	"bytes"
	"errors"
	"fmt"
)

func init() {
	manifest.PluginRegestry.Add("gocd.change", GoCDChange{})
}

type GoCDChange struct{}

/*
plugin for manifest section "gocd.change"
section structure:

gocd.change:
	login: LOGIN
	password: PASSWORD
	url: GOCD_URL
	data:
		group: GROUP
		pipeline:
			according to the description: https://api.go.cd/current/#the-pipeline-config-object

 */
func (p GoCDChange) Run(data manifest.Manifest) error {
	fmt.Println("--> ", data)

	var headers map[string]string
	var name, url string
	body := ""
	cmd := "GET"

	login := data.GetString("login")
	password := data.GetString("password")

	if url = data.GetString("url"); url == "" {
		return errors.New("GoCD url ot found")
	}

	if name = data.GetString("data.pipeline.name"); name == "" {
		return errors.New("GoCD pipeline name not found")
	}

	if resp, err := request(cmd, url + "/" + name, body, headers, login, password); err != nil {
		return err
	} else {
		body = data.String("data")

		if resp.StatusCode == http.StatusOK {
			fmt.Println("put pipeline: ", url)

			cmd = "PUT"
			headers = map[string]string{"If-Match": resp.Header.Get("ETag"), "Content-Type": "application/json"}
			url += "/" + name

		} else {
			fmt.Println("post pipeline ", url)

			cmd = "POST"
			headers = map[string]string{"Content-Type": "application/json"}
		}
	}

	if resp, err := request(cmd, url, body, headers, login, password); err != nil {
		return err
	} else {
		if resp.StatusCode != http.StatusOK {
			return errors.New("Operation error: " + resp.Status)
		}
		return nil
	}
}

func request(	method string,
			resource string,
			body string,
			headers map[string]string,
			login string,
			password string) (*http.Response, error) {

	fmt.Println("method: ", method)
	fmt.Println("resource: ", resource)
	fmt.Println("body", body)

	req, _ := http.NewRequest(method, resource, bytes.NewReader([]byte(body)))
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	req.Header.Set("Accept", "application/vnd.go.cd.v1+json")
	fmt.Println("heads: ", req.Header)
	req.SetBasicAuth(login, password)

	return http.DefaultClient.Do(req)
}
