package plugins

import (
	"github.com/InnovaCo/serve/manifest"
	"net/http"
	"errors"
	"fmt"
)

func init() {
	manifest.PluginRegestry.Add("gocd.delete", GoCDDelete{})
}


type GoCDDelete struct{}

/*
plugin for manifest section "gocd.delete"
section structure:

gocd.delete:
	login: LOGIN
	password: PASSWORD
	url: GOCD_URL
	data:
		group: GROUP
		pipeline:
			name: NAME

 */
func (p GoCDDelete) Run(data manifest.Manifest) error {
	fmt.Println("--> ", data)
	var name, url string

	login := data.GetString("login")
	password := data.GetString("password")

	if url = data.GetString("url"); url == "" {
		return errors.New("GoCD url ot found")
	}

	if name = data.GetString("data.pipeline.name"); name == "" {
		return errors.New("GoCD pipeline name not found")
	}

	if resp, err := request("DELETE", url + "/" + name, "", map[string]string{}, login, password); err != nil {
		return err
	} else {
		if resp.StatusCode == http.StatusOK {
			return nil
		} else {
			errors.New("delete pipeline error: " + resp.Status)
		}
		return nil
	}
}

