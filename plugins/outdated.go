package plugins

import (
	"github.com/servehub/serve/manifest"
	"github.com/servehub/utils"
	"log"
)

func init() {
	manifest.PluginRegestry.Add("outdated", Outdated{})
}

type Outdated struct{}

func (p Outdated) Run(data manifest.Manifest) error {
	if !data.GetBool("enabled") {
		log.Println("Outdated disabled for this service!")
		return nil
	}

	consul, err := utils.ConsulClient(data.GetString("consul-address"))
	if err != nil {
		return err
	}

	fullName := data.GetString("full-name")

	if err := utils.MarkAsOutdated(consul, fullName, 0); err != nil {
		return err
	}

	if err := utils.DelConsulKv(consul, "services/routes/"+fullName); err != nil {
		return err
	}

	return nil
}
