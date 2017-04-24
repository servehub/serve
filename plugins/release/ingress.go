package release

import (
	"fmt"

	"github.com/servehub/serve/manifest"
	"github.com/servehub/serve/plugins/deploy"
)

func init() {
	manifest.PluginRegestry.Add("release.ingress", ReleaseIngress{})
}

type ReleaseIngress struct{}

func (p ReleaseIngress) Run(data manifest.Manifest) error {
	servicePorts := make([]interface{}, 0)
	rules := make([]interface{}, 0)

	data.Set("service.metadata.name", data.GetString("name"))
	data.Set("service.spec.selector.app", data.GetString("app"))

	for _, route := range data.GetArray("routes") {
		servicePorts = append(servicePorts, map[string]interface{}{
			"port": route.GetIntOr("port", 80),
		})

		rules = append(rules, map[string]interface{}{
			"host": route.GetString("host"),
			"http": map[string]interface{}{
				"paths": []interface{}{
					map[string]interface{}{
						"path": route.GetStringOr("path", "/"),
						"backend": map[string]interface{}{
							"serviceName": data.GetString("name"),
							"servicePort": route.GetIntOr("port", 80),
						},
					},
				},
			},
		})
	}

	data.Set("service.spec.ports", servicePorts)

	data.Set("ingress.metadata.name", data.GetString("name"))
	data.Set("ingress.spec.rules", rules)

	if err := deploy.KubeApply("service", data.GetTree("service").Unwrap()); err != nil {
		return fmt.Errorf("Error apply service: %v", err)
	}

	return deploy.KubeApply("ingress", data.GetTree("ingress").Unwrap())
}
