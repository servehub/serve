package plugins

import (
	"fmt"
	"time"

	"github.com/InnovaCo/serve/manifest"
)

func init() {
	manifest.PluginRegestry.Add("outdated", Test{})
}

type Outdated struct{}

func (p Outdated) Run(data manifest.Manifest) error {
	consul, err := ConsulClient(data.GetString("consul-host"))
	if err != nil {
		return err
	}

	fullName := data.GetString("full-name")

	if existsRoutes, _, err := consul.KV().List("services/routes/"+fullName, nil); err == nil {
		for _, existsRoute := range existsRoutes {
			if err := delConsulKv(consul, existsRoute.Key); err != nil {
				return err
			}
		}
	}

	outdatedJson := fmt.Sprintf(`{"endOfLife":%d}`, time.Now().UnixNano()/int64(time.Millisecond))
	if err := putConsulKv(consul, "services/outdated/"+fullName, outdatedJson); err != nil {
		return err
	}

	return nil
}
