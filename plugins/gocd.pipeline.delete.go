package plugins

import (
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
 *  environment: ENV
 */
type GoCdPipelineDelete struct{}

func (p GoCdPipelineDelete) Run(data manifest.Manifest) error {
	if err := goCdDelete(data.GetString("pipeline_name"), data.GetString("environment"),  data.GetString("url")); err != nil {
		log.Println(err)
		return err
	}
	return nil
}
