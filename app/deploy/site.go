package deploy

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	marathon "github.com/gambol99/go-marathon"

	"github.com/InnovaCo/serve/app/build"
	"github.com/InnovaCo/serve/manifest"
	"github.com/fatih/color"
	"github.com/InnovaCo/serve/utils"
)

type SiteDeploy struct {}
type SiteRelease struct {}

func (_ SiteDeploy) Run(m *manifest.Manifest, sub *manifest.Manifest) error {
	conf := marathon.NewDefaultConfig()
	conf.URL = fmt.Sprintf("http://%s:8080", m.GetString("marathon.marathon-host"))
	marathonApi, _ := marathon.NewClient(conf)

	name := m.ServiceName() + "-v" + m.BuildVersion()

	app := &marathon.Application{}
	app.Name(m.GetStringOr("info.category", "") + "/" + name)
	app.Command(fmt.Sprintf("serve consul supervisor --service '%s' --port \\${PORT0} start %s", name, sub.GetStringOr("marathon.cmd", "bin/start")))
	app.Count(sub.GetIntOr("marathon.instances", 1))
	app.Memory(float64(sub.GetIntOr("marathon.mem", 256)))

	if cpu, err := strconv.ParseFloat(sub.GetStringOr("marathon.cpu", "0.1"), 64); err == nil {
		app.CPU(cpu)
	}

	if constrs := sub.GetStringOr("marathon.constraints", ""); constrs != "" {
		cs := strings.SplitN(constrs, ":", 2)
		app.AddConstraint(cs[0], "CLUSTER", cs[1])
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

	// todo: дожидаемся тут появления сервиса в консуле
	return nil
}

func (_ SiteRelease) Run(m *manifest.Manifest, sub *manifest.Manifest) error {
	log.Println("Release done!", sub)

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
