package site

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/cenk/backoff"
	"github.com/fatih/color"
	marathon "github.com/gambol99/go-marathon"

	"github.com/InnovaCo/serve/app/build"
	serveConsul "github.com/InnovaCo/serve/consul"
	"github.com/InnovaCo/serve/manifest"
	serveMarathon "github.com/InnovaCo/serve/marathon"
)

type SiteDeploy struct{}

func (_ SiteDeploy) Run(m *manifest.Manifest, sub *manifest.Manifest) error {
	marathonApi := serveMarathon.MarathonClient(m)

	name := m.ServiceName() + "-v" + m.BuildVersion()

	bs, bf, bmax := 1.0, 2.0, 30.0
	app := &marathon.Application{
		BackoffSeconds:        &bs,
		BackoffFactor:         &bf,
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

	consul := serveConsul.ConsulClient(m)

	return backoff.Retry(func() error {
		services, _, err := consul.Health().Service(name, "", true, nil)

		if err != nil {
			log.Println(color.RedString("Error in check health in consul: %v", err))
			return err
		}

		if len(services) == 0 {
			log.Printf("Service `%s` not started yet! Retry...", name)
			return fmt.Errorf("Service `%s` not started!", name)
		}

		log.Println(color.GreenString("Service `%s` successfully started!", name))
		return nil
	}, backoff.NewExponentialBackOff())
}
