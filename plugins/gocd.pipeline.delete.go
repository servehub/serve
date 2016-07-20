package plugins

import (
	"errors"
	"net/http"
	"github.com/InnovaCo/serve/manifest"
	"log"
)

func init() {
	manifest.PluginRegestry.Add("gocd.pipeline.delete", GoCdPipelineDelete{})
}

/**
 * plugin for manifest section "gocd.delete"
 * section structure:
 *
 * gocd.delete:
 * 	login: LOGIN
 * 	password: PASSWORD
 * 	url: GOCD_URL
 *  pipeline_name: NAME
 */
type GoCdPipelineDelete struct{}

func (p GoCdPipelineDelete) Run(data manifest.Manifest, vars map[string]string) error {
	resp, err := gocdRequest("DELETE", data.GetString("url")+"/"+data.GetString("pipeline_name"), "", nil)
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
