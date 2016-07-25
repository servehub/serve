package plugins

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/cenk/backoff"
	"github.com/fatih/color"
	marathon "github.com/gambol99/go-marathon"
	consul "github.com/hashicorp/consul/api"

	"github.com/InnovaCo/serve/manifest"
)

func init() {
	manifest.PluginRegestry.Add("deploy.marathon", DeployMarathon{})
}

type DeployMarathon struct{}

func (p DeployMarathon) Run(data manifest.Manifest) error {
	marathonApi, err := MarathonClient(data.GetString("marathon-host"))
	if err != nil {
		return err
	}

	fullName := data.GetString("app-name")

	bs, bf, bmax := 1.0, 2.0, 30.0
	app := &marathon.Application{
		BackoffSeconds:        &bs,
		BackoffFactor:         &bf,
		MaxLaunchDelaySeconds: &bmax,
	}

	app.Name(fullName)
	app.Command(fmt.Sprintf("serve-tools consul supervisor --service '%s' --port $PORT0 start %s", fullName, data.GetString("cmd")))
	app.Count(data.GetInt("instances"))
	app.Memory(float64(data.GetInt("mem")))

	if cpu, err := strconv.ParseFloat(data.GetString("cpu"), 64); err == nil {
		app.CPU(cpu)
	}

	if constrs := data.GetString("constraints"); constrs != "" {
		cs := strings.SplitN(constrs, ":", 2)
		app.AddConstraint(cs[0], "CLUSTER", cs[1])
		app.AddLabel(cs[0], cs[1])
	}

	for k, v := range data.GetMap("environment") {
		app.AddEnv(k, fmt.Sprintf("%s", v.Unwrap()))
	}

	app.AddUris(data.GetString("package-uri"))

	if _, err := marathonApi.UpdateApplication(app, false); err != nil {
		color.Yellow("marathon <- %s", app)
		return err
	}

	color.Green("marathon <- %s", app)

	consulApi, err := ConsulClient(data.GetString("consul-host"))
	if err != nil {
		return err
	}

	if err := setKey(consulApi, "/plugins/" + data.GetString("app-name") + "/deploy.marathon", data.String()); err != nil {
		return err
	}

	return backoff.Retry(func() error {
		services, _, err := consulApi.Health().Service(fullName, "", true, nil)

		if err != nil {
			log.Println(color.RedString("Error in check health in consul: %v", err))
			return err
		}

		if len(services) == 0 {
			log.Printf("Service `%s` not started yet! Retry...", fullName)
			return fmt.Errorf("Service `%s` not started!", fullName)
		}

		log.Println(color.GreenString("Service `%s` successfully started!", fullName))
		return nil
	}, backoff.NewExponentialBackOff())
}

func MarathonClient(marathonHost string) (marathon.Marathon, error) {
	conf := marathon.NewDefaultConfig()
	conf.URL = fmt.Sprintf("http://%s:8080", marathonHost)
	conf.LogOutput = os.Stdout
	return marathon.NewClient(conf)
}

func ConsulClient(consulHost string) (*consul.Client, error) {
	conf := consul.DefaultConfig()
	conf.Address = consulHost + ":8500"
	return consul.NewClient(conf)
}

func setKey(client *consul.Client, key string, value string) error {
	kv := client.KV()
	p := &consul.KVPair{Key: key, Value: []byte(value)}
	if _, err := kv.Put(p, nil); err != nil {
		return err
	}
	return nil
}