package supervisor

import (
	"log"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"github.com/cenk/backoff"
	"github.com/codegangsta/cli"
	"github.com/fatih/color"
)

func SupervisorCommand() cli.Command {
	return cli.Command{
		Name:            "supervisor",
		SkipFlagParsing: true,
		Action: func(c *cli.Context) error {
			var cmd *exec.Cmd = nil

			go func() {
				backoff.Retry(func() error {
					log.Println(color.GreenString("supervisor: Starting %v", c.Args()))

					cmd = exec.Command(c.Args().First(), c.Args().Tail()...)
					cmd.Stdout = os.Stdout
					cmd.Stderr = os.Stderr

					if err := cmd.Start(); err != nil {
						log.Println(color.RedString("supervisor: Error on process staring %v", err))
						return err
					}

					err := cmd.Wait()
					if err != nil {
						log.Println(color.RedString("supervisor: Command exit with error: %v", err))
					} else {
						log.Println(color.YellowString("supervisor: Command completed."))
						os.Exit(0)
					}
					return err
				}, backoff.NewConstantBackOff(time.Second * 3))
			}()

			// Handle shutdown signals and kill child process
			ch := make(chan os.Signal)
			signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)
			log.Println("supervisor: signal", <-ch)

			if cmd != nil {
				cmd.Process.Signal(syscall.SIGTERM)
				time.Sleep(time.Second) // await while child stopped
			}

			log.Println(color.YellowString("supervisor: Stopped."))
			return nil
		},
	}
}
