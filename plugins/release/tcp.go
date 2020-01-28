package release

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/cenk/backoff"
	"github.com/fatih/color"

	"github.com/servehub/serve/manifest"
	"github.com/servehub/utils"
)

func init() {
	manifest.PluginRegestry.Add("release.tcp", ReleaseTcp{})
}

type ReleaseTcp struct{}

func (p ReleaseTcp) Run(data manifest.Manifest) error {
	if !data.Has("port") {
		log.Println("No port configured for tcp release.")
		return nil
	}

	consul, err := utils.ConsulClient(data.GetString("consul-address"))
	if err != nil {
		return err
	}

	fullName := data.GetString("full-name")

	// check current service is alive and healthy
	if err := backoff.Retry(func() error {
		services, _, err := consul.Health().Service(fullName, "", true, nil)
		if err != nil {
			log.Println(color.RedString("Error in check health in consul: %v", err))
			return err
		}

		if len(services) == 0 {
			log.Printf("Service `%s` not started yet! Retry...", fullName)
			return fmt.Errorf("Service `%s` not started!", fullName)
		} else {
			log.Printf("Service `%s` started with %v instances.", fullName, len(services))
			return nil
		}
	}, backoff.NewExponentialBackOff()); err != nil {
		return err
	}

	routeVars := make(map[string]string, 1)
	if data.Has("public") {
		routeVars["public"] = data.GetString("public")
	}

	tcpJson, err := json.MarshalIndent(consulTcpRoute{
		Port:     data.GetInt("port"),
		Protocol: data.GetString("protocol"),
		Vars:     routeVars,
	}, "", "  ")

	if err != nil {
		return err
	}

	// write tcp routes to consul kv
	if err := utils.PutConsulKv(consul, "services/tcp-routes/"+fullName, string(tcpJson)); err != nil {
		return err
	}

	log.Println(color.GreenString("Service `%s` released with tcp: %s", fullName, string(tcpJson)))

	return nil
}

type consulTcpRoute struct {
	Port     int               `json:"port"`
	Protocol string            `json:"protocol"`
	Vars     map[string]string `json:"vars,omitempty"`
}
