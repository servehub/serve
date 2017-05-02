package consul

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/codegangsta/cli"
	"github.com/hashicorp/consul/api"
)

func SupervisorCommand() cli.Command {
	return cli.Command{
		Name:  "supervisor",
		Usage: "Wrapper for registering service into consul and remove it after stop",
		Flags: []cli.Flag{
			cli.StringFlag{Name: "service"},
			cli.StringFlag{Name: "port"},
		},
		Subcommands: []cli.Command{
			{
				Name:            "start",
				SkipFlagParsing: true,
				Action: func(c *cli.Context) {
					log.Println("Starting", c.Args())

					cmd := exec.Command(c.Args().First(), c.Args().Tail()...)
					cmd.Stdout = os.Stdout
					cmd.Stderr = os.Stderr

					if err := cmd.Start(); err != nil {
						log.Fatal("Error on process staring", err)
					}

					consul, _ := api.NewClient(api.DefaultConfig())

					var serviceId string
					var serviceReg *api.AgentServiceRegistration

					if c.GlobalString("port") != "" {
						port, err := strconv.Atoi(c.GlobalString("port"))
						if err != nil {
							log.Fatalf("Error on get service --port %s: %s", c.GlobalString("port"), err)
						}

						serviceId = fmt.Sprintf("%s:%d", c.GlobalString("service"), port)

						serviceReg = &api.AgentServiceRegistration{
							ID:   serviceId,
							Name: c.GlobalString("service"),
							Port: port,
							Check: &api.AgentServiceCheck{
								TCP:                            "localhost:" + strconv.Itoa(port),
								Interval:                       "10s",
								DeregisterCriticalServiceAfter: "60m",
							},
						}
					} else {
						serviceId = fmt.Sprintf("%s:%d", c.GlobalString("service"), time.Now().UnixNano())

						serviceReg = &api.AgentServiceRegistration{
							ID:   serviceId,
							Name: c.GlobalString("service"),
						}
					}

					// wait for child process compelete and unregister it from consul
					go func() {
						result := cmd.Wait()
						log.Printf("Service finished with: %v", result)

						log.Println("Deregister service", serviceId, "...")
						if err := consul.Agent().ServiceDeregister(serviceId); err != nil {
							log.Fatal("Error on deregistering service", err)
						}

						log.Println("Successful deregistered.")

						if exiterr, ok := result.(*exec.ExitError); ok {
							if status, ok := exiterr.Sys().(syscall.WaitStatus); ok && status.Exited() {
								os.Exit(status.ExitStatus())
							}
						}

						if result != nil {
							os.Exit(143)
						} else {
							os.Exit(0)
						}
					}()

					// Register service into consul
					if err := consul.Agent().ServiceRegister(serviceReg); err != nil {
						cmd.Process.Signal(syscall.SIGTERM)
						log.Fatal(err)
					}

					// Handle shutdown signals and kill child process
					ch := make(chan os.Signal, 1)
					signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
					log.Println("Receive signal", <-ch)

					cmd.Process.Signal(syscall.SIGTERM)

					result, err := cmd.Process.Wait()
					log.Printf("Command finished with: %v, %v", result, err)

					time.Sleep(10 * time.Second) // await while service deregistered from consul

					log.Println("Stopped.")
				},
			},
		},
	}
}
