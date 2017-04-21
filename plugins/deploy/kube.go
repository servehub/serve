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

		for _, p := range data.GetArray("ports") {
			cnt.Set("livenessProbe.tcpSocket.port", p.GetInt("containerPort"))
		}
	}

	deployment, err := yaml.Marshal(data.GetTree("deployment").Unwrap())
	if err != nil {
		return fmt.Errorf("Error on serialize deployment yaml: %v", err)
	}

	tmpfile, err := ioutil.TempFile("", "deployment")
	if err != nil {
		return fmt.Errorf("Error create tmpfile: %v", err)
	}

	defer func() {
		tmpfile.Close()
		os.Remove(tmpfile.Name())
	}()

	if _, err := tmpfile.Write(deployment); err != nil {
		return fmt.Errorf("Error write to tmpfile: %v", err)
	}

	return utils.RunCmd("kubectl apply -o yaml -f %s", tmpfile.Name())
}
