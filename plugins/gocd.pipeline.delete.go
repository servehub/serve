package plugins

import (
	"errors"
	"net/http"

	"github.com/InnovaCo/serve/manifest"
)

func init() {
	manifest.PluginRegestry.Add("gocd.pipeline.delete", GoCdPipelineDelete{})
}

/**
 * plugin for manifest section "gocd.delete"
 * section structure:
 *
 * gocd.pipeline.create:
 * 	login: LOGIN
 * 	password: PASSWORD
 * 	url: GOCD_URL
 * 	data:
 * 		group: GROUP
 * 		pipeline:
 * 			name: NAME
 */
type GoCdPipelineDelete struct{}

func (p GoCdPipelineDelete) Run(data manifest.Manifest) error {
	resp, err := gocdRequest("DELETE", data.GetString("url")+"/"+data.GetString("pipeline_name"), "", nil)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return errors.New("Operation error: " + resp.Status)
	}

	return nil
}
