package deploy

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/cenk/backoff"
	"github.com/fatih/color"
	marathon "github.com/gambol99/go-marathon"
	"github.com/hashicorp/consul/api"

	"github.com/InnovaCo/serve/app/build"
	"github.com/InnovaCo/serve/manifest"
	"github.com/InnovaCo/serve/utils"
)

type SiteDeploy struct{}
type SiteRelease struct{}

func (_ SiteDeploy) Run(m *manifest.Manifest, sub *manifest.Manifest) error {
	conf := marathon.NewDefaultConfig()
	conf.URL = fmt.Sprintf("http://%s:8080", m.GetString("marathon.marathon-host"))
	marathonApi, _ := marathon.NewClient(conf)

	name := m.ServiceName() + "-v" + m.BuildVersion()

	bs, bf, bmax := 1.0, 2.0, 30.0
	app := &marathon.Application{
		BackoffSeconds: &bs,
		BackoffFactor: &bf,
		MaxLaunchDelaySeconds: &bmax,
	}

	app.Name(m.GetStringOr("info.category", "") + "/" + name)
	app.Command(fmt.Sprintf("serve consul supervisor --service '%s' --port $PORT0 start %s", name, sub.GetStringOr("marathon.cmd", "bin/start")))
	app.Count(sub.GetIntOr("marathon.instances", 1))
	app.Memory(float64(sub.GetIntOr("marathon.mem", 256)))

	if cpu, err := strconv.ParseFloat(sub.GetStringOr("marathon.cpu", "0.1"), 64); err == nil {
		app.CPU(cpu)
	}

	if constrs := sub.GetStringOr("marathon.constraints", ""); constrs != "" {
		cs := strings.SplitN(constrs, ":", 2)
		app.AddConstraint(cs[0], "CLUSTER", cs[1])
		app.AddLabel(cs[0], cs[1])
	}

	app.AddEnv("ENV", m.Args("env"))
	app.AddEnv("SERVICE_NAME", m.ServiceName())
	app.AddEnv("MEMORY", sub.GetStringOr("marathon.mem", ""))

	app.AddUris(build.TaskRegistryUrl(m))

	if _, err := marathonApi.UpdateApplication(app, false); err != nil {
		color.Yellow("marathon <- %s", app)
		return err
	}

	color.Green("marathon <- %s", app)

	consul := ConsulClient(m)

	return backoff.Retry(func() error {
		services, _, err := consul.Health().Service(name, "", true, nil)

		if err != nil {
			return err
		}

		if len(services) == 0 {
			log.Printf("Service `%s` not started yet! Retry...", name)
			return fmt.Errorf("Service `%s` not started!", name)
		}

		log.Println(color.GreenString("Service `%s` successfily deloyed!", name))
		return nil
	}, backoff.NewExponentialBackOff())
}

func (_ SiteRelease) Run(m *manifest.Manifest, sub *manifest.Manifest) error {
	log.Println("Release done!", sub)

	conf := api.DefaultConfig()
	conf.Address = m.GetString("consul.consul-host") + ":8500"

	consul, _ := api.NewClient(conf)

	routes := make([]map[string]string, 0)
	for _, route := range sub.Array("routes") {

		// todo: merge with --route flag
		// filter featured: true route
		routes = append(routes, map[string]string{
			"host":     route.GetString("host"),
			"location": route.GetString("location"),
		})
	}

	consul.KV().Put(&api.KVPair{
		Key:   fmt.Sprintf("services/%s/%s/routes", m.ServiceName(), m.BuildVersion()),
		Value: []byte("test"),
	}, nil)

	// находим текущий в консуле и убеждаемся что с ним все ок
	// добавляем ему роуты

	// ищем есть ли старый с такими же роутами:
	//    формируем массив роутов
	//    ищем сервис с таким-же именем но другой версии, и содержащий один из указанных роутов
	//    например в kv можно хранить /kv/services/{name-?branch}/v{version}} и там матчить через compareMaps
	//    если хотябы один роут полностью совпал — это наш кандидат на убивание
	// если есть — убиваем в консуле сразу и через 5 минут в марафоне

	println(utils.MapsEqual(map[string]string{"name": "dima", "version": "1.0"}, map[string]string{"version": "1.0", "name": "dima"}))

	log.Println("route")

	return nil
}

func ConsulClient(m *manifest.Manifest) *api.Client {
	conf := api.DefaultConfig()
	conf.Address = m.GetString("consul.consul-host") + ":8500"

	consul, _ := api.NewClient(conf)
	return consul
}
