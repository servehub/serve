package plugins

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

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
	if data.GetBool("purge") {
		return p.Uninstall(data)
	} else {
		return p.Install(data)
	}
}

func (p DeployMarathon) Install(data manifest.Manifest) error {
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

	portArgs := ""
	if port := data.GetString("listen-port"); port != "" {
		portArgs = "--port " + port
	}

	app.Name(fullName)
	app.Command(fmt.Sprintf("exec serve-tools consul supervisor --service '%s' %s start %s", fullName, portArgs, data.GetString("cmd")))
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

	app.AddEnv("SERVICE_DEPLOY_TIME", time.Now().Format(time.RFC3339)) // force redeploy app

	for k, v := range data.GetMap("environment") {
		app.AddEnv(k, fmt.Sprintf("%s", v.Unwrap()))
	}

	app.AddUris(data.GetString("package-uri"))

	// todo: в манифесте задавать массив healthchecks, их использовтаь в марафоне и консул-супервизоре
	// todo: открыть сетевой доступ от марафона до мезос-агентов, чтобы марафон мог хелсчеки посылать

	//if portArgs != "" {
	//	health := marathon.NewDefaultHealthCheck()
	//	health.Protocol = "TCP"
	//	health.IntervalSeconds = 5
	//	*health.PortIndex = 0
	//	app.AddHealthCheck(*health)
	//}

	if _, err := marathonApi.UpdateApplication(app, false); err != nil {
		color.Yellow("marathon <- %s", app)
		return err
	}

	color.Green("marathon <- %s", app)

	consulApi, err := ConsulClient(data.GetString("consul-address"))
	if err != nil {
		return err
	}

	if err := registerPluginData("deploy.marathon", data.GetString("app-name"), data.String(), data.GetString("consul-address")); err != nil {
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

func (p DeployMarathon) Uninstall(data manifest.Manifest) error {
	marathonApi, err := MarathonClient(data.GetString("marathon-host"))
	if err != nil {
		return err
	}

	name := data.GetString("app-name")

	if _, err := marathonApi.Application(name); err == nil {
		if _, err := marathonApi.DeleteApplication(name, false); err != nil {
			return err
		}
	} else {
		log.Println(color.YellowString("App `%s` doesnt exists in marathon!", name))
	}

	return deletePluginData("deploy.marathon", name, data.GetString("consul-address"))
}

func MarathonClient(marathonHost string) (marathon.Marathon, error) {
	conf := marathon.NewDefaultConfig()
	conf.URL = fmt.Sprintf("http://%s:8080", marathonHost)
	conf.LogOutput = os.Stderr
	return marathon.NewClient(conf)
}

func ConsulClient(consulAddress string) (*consul.Client, error) {
	conf := consul.DefaultConfig()
	conf.Address = consulAddress
	return consul.NewClient(conf)
}

func putConsulKv(client *consul.Client, key string, value string) error {
	log.Printf("consul put `%s`: %s", key, value)
	_, err := client.KV().Put(&consul.KVPair{Key: strings.TrimPrefix(key, "/"), Value: []byte(value)}, nil)
	return err
}

func listConsulKv(client *consul.Client, prefix string, q *consul.QueryOptions) (consul.KVPairs, error) {
	log.Printf("consul list `%s`", prefix)
	list, _, err := client.KV().List(prefix, q)
	return list, err
}

func delConsulKv(client *consul.Client, key string) error {
	log.Printf("consul delete `%s`", key)
	_, err := client.KV().Delete(strings.TrimPrefix(key, "/"), nil)
	return err
}

func registerPluginData(plugin string, packageName string, data string, consulAddress string) error {
	consulApi, err := ConsulClient(consulAddress)
	if err != nil {
		return err
	}

	return putConsulKv(consulApi, "services/data/"+packageName+"/"+plugin, data)
}

func deletePluginData(plugin string, packageName string, consulAddress string) error {
	log.Println(color.YellowString("Delete %s for %s package in consul", plugin, packageName))
	consulApi, err := ConsulClient(consulAddress)
	if err != nil {
		return err
	}

	return delConsulKv(consulApi, "services/data/"+packageName+"/"+plugin)
}

func markAsOutdated(client *consul.Client, name string, delay time.Duration) error {
	log.Printf("Mark service `%s` as outdated\n", name)
	json := fmt.Sprintf(`{"endOfLife":%d}`, time.Now().Add(delay).UnixNano()/int64(time.Millisecond))
	return putConsulKv(client, "services/outdated/"+name, json)
}
