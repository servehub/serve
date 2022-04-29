package gocd

import (
	"github.com/servehub/serve/manifest"
	"github.com/servehub/utils"
)

func init() {
	manifest.PluginRegestry.Add("gocd.pipeline.run", goCdPipelineRun{})
}

type goCdPipelineRun struct{}

func (p goCdPipelineRun) Run(data manifest.Manifest) error {
	url := data.GetString("api-url")
	name := data.GetString("pipeline-name")

	resp, err := goCdSchedule(name, url, data.GetTree("schedule").String(), map[string]string{"Accept": "application/vnd.go.cd.v1+json"})

	if err == nil && resp.StatusCode == 404 {
		if err := utils.RunCmd(
			"serve gocd.pipeline.create --ssh-repo=$GO_MATERIAL_URL_SOURCES --branch=%s",
			data.GetString("branch"),
		); err != nil {
			return err
		}

		_, err := goCdSchedule(name, url, data.GetTree("schedule").String(), map[string]string{"Accept": "application/vnd.go.cd.v1+json"})
		return err
	}

	return err
}
