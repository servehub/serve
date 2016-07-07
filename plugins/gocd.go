package plugins

import (
	"github.com/InnovaCo/serve/manifest"
	"net/http"
	"bytes"
	"errors"
	"fmt"
)

func init() {
	manifest.PluginRegestry.Add("gocd", GoCD{})
}

type GoCD struct{}

/*
function for GoCD actions: ADD/EDIT/DELETE
tree is structure:
	login: "LOGIN"
	password: "PASSWORD"
	url: "URL"
	delete: "TRUE/FALSE"
	pipeline: PARAMS
 */
func (p GoCD) Run(data manifest.Manifest) error {
	fmt.Println("--> ", data)

	login := data.GetString("login")
	password := data.GetString("password")

	url := data.GetString("url")
	if url == "" {
		return errors.New("GoCD url ot found")
	}

	del := data.GetString("delete")
	if del == "" {
		del = "false"
	}

	name := data.GetString("data.pipeline.name")
	if name == "" {
		return errors.New("GoCD pipeline name not found")
	}

	var headers map[string]string
	body := ""
	cmd := "GET"

	if del == "true"{
		fmt.Println("delete pipeline: ", url)

		cmd = "DELETE"
		resp, err := p.request(cmd, url, body, headers, login, password)
		if err != nil {
			return err
		}

		if resp.StatusCode == http.StatusOK {
			return nil
		} else {
			errors.New("delete pipeline error: " + resp.Status)
		}
		return nil
	}

	resp, err := p.request(cmd, url + "/" + name, body, headers, login, password)
	if err != nil {
		return err
	}

	body = data.GenString("data")

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

	resp, err = p.request(cmd, url, body, headers, login, password)

	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return errors.New("Operation error: " + resp.Status)
	}
	return nil
}

func (p GoCD) request(	method string,
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
