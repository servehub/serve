package deploy

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/ghodss/yaml"

	"github.com/servehub/serve/manifest"
	"github.com/servehub/utils"
)

func init() {
	manifest.PluginRegestry.Add("deploy.kube", DeployKube{})
}

type DeployKube struct{}

func (p DeployKube) Run(data manifest.Manifest) error {
	envsMap := make(map[string]interface{}, 0)
	envs := make([]map[string]interface{}, 0)

	for k, v := range data.GetMap("envs") {
		envsMap[k] = v.Unwrap()
	}
	for k, v := range data.GetMap("environment") {
		envsMap[k] = v.Unwrap()
	}

	envsMap["SERVICE_DEPLOY_TIME"] = time.Now().Format(time.RFC3339) // force redeploy app
	envsMap["SERVICE_MEMORY"] = strings.ToLower(data.GetString("requests.memory"))

	for k, v := range envsMap {
		if m, ok := v.(map[string]interface{}); ok {
			m["name"] = k
			envs = append(envs, m)
		} else {
			envs = append(envs, map[string]interface{}{
				"name":  k,
				"value": strings.TrimSpace(fmt.Sprintf("%v", v)),
			})
		}
	}

	data.Set("deployment.spec.replicas", data.GetInt("deployment.spec.replicas"))

	for _, cnt := range data.GetArray("deployment.spec.template.spec.containers") {
		cnt.Set("env", envs)
		cnt.Set("ports", data.GetTree("ports").Unwrap())

		cnt.Set("resources.requests", data.GetTree("requests").Unwrap())
		cnt.Set("resources.limits", data.GetTree("limits").Unwrap())

		for _, p := range data.GetArray("ports") {
			cnt.Set("readinessProbe", data.GetTree("readinessProbe").Unwrap())
			cnt.Set("readinessProbe.tcpSocket.port", p.GetInt("containerPort"))
			break // check only first port
		}
	}

	return KubeApply("deployment", data.GetTree("deployment").Unwrap())
}

func KubeApply(name string, data interface{}) error {
	bytes, err := yaml.Marshal(data)
	if err != nil {
		return fmt.Errorf("Error on serialize %s yaml: %v", name, err)
	}

	tmpfile, err := ioutil.TempFile("", "serve-kube-"+name)
	if err != nil {
		return fmt.Errorf("Error create tmpfile: %v", err)
	}

	defer func() {
		tmpfile.Close()
		os.Remove(tmpfile.Name())
	}()

	if _, err := tmpfile.Write(bytes); err != nil {
		return fmt.Errorf("Error write to tmpfile: %v", err)
	}

	return utils.RunCmd("kubectl apply -o yaml -f %s", tmpfile.Name())
}
