package plugins

import (
	"github.com/InnovaCo/serve/manifest"
)

func init() {
	manifest.PluginRegestry.Add("gocd.pipeline.delete", GoCdPipelineDelete{})
}

/**
 * plugin for manifest section "gocd.delete"
 * section structure:
 *
 * gocd.delete:
 *   api-url: GOCD_URL
 *   environment: ENV
 *   pipeline
 *     name: NAME
 */
type GoCdPipelineDelete struct{}

func (p GoCdPipelineDelete) Run(data manifest.Manifest) error {
	return goCdDelete(data.GetString("pipeline.name"), data.GetString("environment"), data.GetString("api-url"),
		              map[string]string{"Accept": "application/vnd.go.cd.v2+json"})
}
