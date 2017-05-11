package deploy

import (
	"encoding/json"

	"github.com/servehub/serve/manifest"
	"github.com/servehub/utils"
)

func init() {
	manifest.PluginRegestry.Add("deploy.secrets", DeploySecrets{})
}

type DeploySecrets struct{}

func (p DeploySecrets) Run(data manifest.Manifest) error {
	env := data.GetString("env")
	data.DelTree("env")

	consul, err := utils.ConsulClient(data.GetString("consul.address"))
	consulPath := data.GetString("consul.path")
	if err != nil {
		return err
	}
	data.DelTree("consul")

	for key, sec := range data.GetMap(".") {
		if sec.Has("value." + env) {
			sec.Set("value", sec.GetString("value." + env))
		} else {
			if _, ok := sec.GetTree("value").Unwrap().(string); !ok {
				data.DelTree(key)
			}
		}
	}

	body, _ := json.MarshalIndent(map[string]interface{}{"secrets": data.Unwrap()}, "", "  ")
	return utils.PutConsulKv(consul, consulPath, string(body))
}
