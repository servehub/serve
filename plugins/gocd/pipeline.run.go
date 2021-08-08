package gocd

import (
	"github.com/servehub/serve/manifest"
)

func init() {
	manifest.PluginRegestry.Add("gocd.pipeline.run", goCdPipelineRun{})
}

type goCdPipelineRun struct{}

func (p goCdPipelineRun) Run(data manifest.Manifest) error {
	url := data.GetString("api-url")
	name := data.GetString("pipeline-name")

	return goCdSchedule(name, url, data.GetTree("schedule").String(), map[string]string{"Accept": "application/vnd.go.cd.v1+json"})
}
