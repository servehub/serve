package plugins

import (
	"log"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"github.com/hashicorp/consul/api"

	"github.com/InnovaCo/serve/manifest"
	"flag"
	"strconv"
)

func init() {
	manifest.PluginRegestry.Add("consul.supervisor", ConsulSupervisor{})
}

type ConsulSupervisor struct{}

func (p ConsulSupervisor) Run(data manifest.Manifest) error {
	service := *flag.String("service", "", "")
	port := *flag.Int("port", 0, "")

	log.Println(service)
	log.Println(port)
	log.Println(flag.CommandLine.Args())

	log.Println("Starting", data.GetString("command"))

	cmd := exec.Command(flag.CommandLine.Args()[0], flag.CommandLine.Args()[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		log.Fatal("Error on process staring", err)
	}

	consul, _ := api.NewClient(api.DefaultConfig())

	serviceId := service + ":" + strconv.Itoa(port)

	// wait for child process compelete and unregister it from consul
	go func() {
		result := cmd.Wait()
		log.Printf("Command finished with: %v", result)

		log.Println("Deregister service", serviceId, "...")
		if err := consul.Agent().ServiceDeregister(serviceId); err != nil {
			log.Fatal(err)
		}

		log.Println("Deregistered.")

		if exiterr, ok := result.(*exec.ExitError); ok {
			if status, ok := exiterr.Sys().(syscall.WaitStatus); ok && status.Exited() {
				os.Exit(status.ExitStatus())
			}
		}

		if result != nil {
			os.Exit(2)
		} else {
			os.Exit(0)
		}
	}()

	// Register service into consul
	if err := consul.Agent().ServiceRegister(&api.AgentServiceRegistration{
		ID:   serviceId,
		Name: service,
		Port: port,
		Check: &api.AgentServiceCheck{
			TCP:      "localhost:" + strconv.Itoa(port),
			Interval: "5s",
		},
	}); err != nil {
		cmd.Process.Signal(syscall.SIGTERM)
		log.Fatal(err)
	}

	// Handle shutdown signals and kill child process
	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)
	log.Println(<-ch)

	cmd.Process.Signal(syscall.SIGTERM)
	time.Sleep(time.Second) // await while child stopped

	log.Println("Stopped.")

	return nil
}
