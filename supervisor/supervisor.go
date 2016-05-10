package supervisor

import (
	"log"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"
)

type Supervisor struct {
	MaxRetry   int
	OnError    func()
	OnComplete func()
}

func (s Supervisor) Run(cmdArgs []string) {
	cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		log.Fatal("Error on process staring", err)
	}

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
			os.Exit(0)
		} else {
			os.Exit(2)
		}
	}()

	// Register service into consul
	if err := consul.Agent().ServiceRegister(&api.AgentServiceRegistration{
		ID:                serviceId,
		Name:              c.GlobalString("name"),
		Tags:              MapToList(TagsFromFlags(c)),
		Port:              c.GlobalInt("port"),
		EnableTagOverride: true,
		Check: &api.AgentServiceCheck{
			TCP:      "localhost:" + c.GlobalString("port"),
			Interval: "5s",
		},
	}); err != nil {
		cmd.Process.Kill()
		log.Fatal(err)
	}

	// Handle shutdown signals and kill child process
	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)
	log.Println(<-ch)

	cmd.Process.Kill()
	time.Sleep(time.Second)
}
