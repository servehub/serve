package utils

import (
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/fatih/color"
)

func RunCmd(cmdline string, a ...interface{}) error {
	return RunCmdWithEnv(fmt.Sprintf(cmdline, a...), make(map[string]string, 0))
}

var RunCmdWithEnv = func(cmdline string, env map[string]string) error {
	log.Println(color.YellowString("> %s", cmdline))

	cmd := exec.Command("/bin/bash", "-c", cmdline)

	cmd.Env = os.Environ()

	for k, v := range env {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%v", k, v))
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func RunSshCmd(cluster, sshUser, cmd string) error {
	return RunParallelSshCmd(cluster, sshUser, cmd, 1)
}

func RunParallelSshCmd(cluster, sshUser, cmd string, maxProcs int) error {
	if cluster == "" {
		return fmt.Errorf("RunParallelSshCmd: `cluster` must not be empty! Cmd: %s", cmd)
	}

	return RunCmd(
		`dig +short %s | sort | uniq | parallel --tag --line-buffer -j %d ssh -i ~/.ssh/id_rsa -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null %s@{} "%s"`,
		cluster,
		maxProcs,
		sshUser,
		cmd,
	)
}

func RunSingleSshCmd(cluster, sshUser, cmd string) error {
	if cluster == "" {
		return fmt.Errorf("RunSingleSshCmd: `cluster` must not be empty! Cmd: %s", cmd)
	}

	return RunCmd(
		`ssh -i ~/.ssh/id_rsa -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null %s@%s "%s"`,
		sshUser,
		cluster,
		cmd,
	)
}
