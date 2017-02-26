package deploy

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

	"github.com/servehub/serve/manifest"
	"github.com/servehub/utils"
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
	marathonApi, err := MarathonClient(data.GetString("marathon-address"))
	if err != nil {
		return err
	}

	fullName := data.GetString("app-name")

	backoffSeconds := data.GetFloat("backoff-seconds")
	backoffFactor := data.GetFloat("backoff-factor")
	maxLaunchDelaySeconds := data.GetFloat("max-launch-delay-seconds")
	taskKillGracePeriodSeconds := data.GetFloat("task-kill-grace-period-seconds")

	app := &marathon.Application{
		User:                       data.GetString("user"),
		BackoffSeconds:             &backoffSeconds,
		BackoffFactor:              &backoffFactor,
		MaxLaunchDelaySeconds:      &maxLaunchDelaySeconds,
		TaskKillGracePeriodSeconds: &taskKillGracePeriodSeconds,
		UpgradeStrategy: &marathon.UpgradeStrategy{
			MinimumHealthCapacity: data.GetFloat("min-health-capacity"),
			MaximumOverCapacity:   data.GetFloat("max-over-capacity"),
		},
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

	if cluster := data.GetString("cluster"); cluster != "" {
		cs := strings.SplitN(cluster, ":", 2)
		if len(cs) < 2 {
			cs = append(cs, "true")
		}
		app.AddConstraint(cs[0], "CLUSTER", cs[1])
		app.AddLabel(cs[0], cs[1])
	}

	for _, cons := range data.GetArray("constraints") {
		if consArr, ok := cons.Unwrap().([]interface{}); ok {
			consStrings := make([]string, len(consArr))
			for i, c := range consArr {
				consStrings[i] = fmt.Sprintf("%s", c)
			}
			app.AddConstraint(consStrings...)
		}
	}

	for _, port := range data.GetArray("ports") {
		app.AddPortDefinition(marathon.PortDefinition{Name: port.GetStringOr("name", "")}.SetPort(port.GetIntOr("port", 0)))
	}

	app.AddEnv("SERVICE_DEPLOY_TIME", time.Now().Format(time.RFC3339)) // force redeploy app

	for k, v := range data.GetMap("environment") {
		app.AddEnv(k, fmt.Sprintf("%s", v.Unwrap()))
	}

	app.AddUris(data.GetString("package-uri"))

	if data.GetBool("docker.enabled") {
		app.Cmd = nil
		app.EmptyUris()
		app.EmptyPortDefinitions()

		doc := marathon.NewDockerContainer()
		doc.Docker.Image = data.GetString("docker.image")
		doc.Docker.Network = data.GetString("docker.network")
		doc.Docker.SetForcePullImage(false) // todo: use true by default
		doc.EmptyVolumes()

		for _, port := range data.GetArray("docker.ports") {
			doc.Docker.ExposePort(marathon.PortMapping{
				ContainerPort: port.GetInt("containerPort"),
				HostPort:      port.GetIntOr("hostPort", 0),
				Protocol:      "tcp",
			})
		}

		for _, vol := range data.GetArray("docker.volumes") {
			doc.Volume(vol.GetString("hostPath"), vol.GetString("containerPath"), vol.GetString("mode"))
		}

		app.Container = doc
	}

	if data.GetString("listen-port") == "$PORT0" {
		health := marathon.NewDefaultHealthCheck()
		health.Protocol = "TCP"
		app.AddHealthCheck(*health)
	}

	if _, err := marathonApi.UpdateApplication(app, false); err != nil {
		color.Yellow("marathon <- %s", app)
		return err
	}

	color.Green("marathon <- %s", app)

	consulApi, err := utils.ConsulClient(data.GetString("consul-address"))
	if err != nil {
		return err
	}

	if err := utils.RegisterPluginData("deploy.marathon", fullName, data.String(), data.GetString("consul-address")); err != nil {
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
	marathonApi, err := MarathonClient(data.GetString("marathon-address"))
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

	return utils.DeletePluginData("deploy.marathon", name, data.GetString("consul-address"))
}

func MarathonClient(marathonAddress string) (marathon.Marathon, error) {
	conf := marathon.NewDefaultConfig()
	conf.URL = fmt.Sprintf("http://%s", marathonAddress)
	conf.LogOutput = os.Stderr
	return marathon.NewClient(conf)
}
