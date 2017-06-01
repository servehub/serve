package monitoring

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/servehub/serve/manifest"
	"github.com/servehub/utils"
)

func init() {
	manifest.PluginRegestry.Add("monitoring", MonitoringRun{})
}

var nameRegepx = regexp.MustCompile(`[^a-z0-9-]`)

type MonitoringRun struct{}

func (p MonitoringRun) Run(data manifest.Manifest) error {
	consul, err := utils.ConsulClient(data.GetString("consul.address"))
	consulPath := data.GetString("consul.path")
	if err != nil {
		return err
	}

	alerts := make(map[string]map[string]interface{}, 0)

	for key, alert := range data.GetMap("alerts") {
		if alert.Has("elastic") {
			if _, ok := alerts["elastic"]; !ok {
				alerts["elastic"] = make(map[string]interface{}, 0)
			}

			name := nameRegepx.ReplaceAllString(strings.ToLower(key), "-")

			alerts["elastic"][name] = map[string]interface{}{
				"name":  key,
				"query": alert.GetString("elastic"),
			}
		}
	}

	for alertType, alert := range alerts {
		body, _ := json.MarshalIndent(alert, "", "  ")

		if err := utils.PutConsulKv(consul, fmt.Sprintf(consulPath, alertType), string(body)); err != nil {
			return err
		}
	}

	return nil
}
