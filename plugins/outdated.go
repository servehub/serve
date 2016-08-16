package plugins

import (
	"github.com/InnovaCo/serve/manifest"
)

func init() {
	manifest.PluginRegestry.Add("outdated", Outdated{})
}

type Outdated struct{}

func (p Outdated) Run(data manifest.Manifest) error {
	consul, err := ConsulClient(data.GetString("consul-host"))
	if err != nil {
		return err
	}

	fullName := data.GetString("full-name")

	if err := markAsOutdated(consul, fullName, 0); err != nil {
		return err
	}

	if err := delConsulKv(consul, "services/routes/"+fullName); err != nil {
		return err
	}

	return nil
}
