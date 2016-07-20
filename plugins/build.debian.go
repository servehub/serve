package plugins

import (
	"github.com/InnovaCo/serve/manifest"
	"fmt"
	"strings"
	"os/exec"
)

const ciToolsPath = "/var/go/inn-ci-tools"

func init() {
	manifest.PluginRegestry.Add("build.debian", BuildDebian{})
}

type BuildDebian struct {}

func(p BuildDebian) Run(data manifest.Manifest, vars map[string]string) error {
	nameWithVersion := data.GetString("name_version")
	name := data.GetString("name")
	daemonArgs := data.GetString("daemon_args")
	serviceOwner := data.GetString("service_owner")
	daemonUser := data.GetString("daemon_user")
	daemon := data.GetString("daemon")
	daemonPort := data.GetString("daemon_port")
	makePidfile := data.GetString("make_pidfile")
	installRoot := data.GetString("install_root")
	depends := data.GetString("depends")
	version := data.GetString("version")
	maintainerName := data.GetString("maintainer_name")
	maintainerEmail := data.GetString("maintainer_email")
	section := data.GetString("section")
	description := data.GetString("description")
	category := data.GetString("category")
	init := data.GetString("init")
	cron := data.GetString("cron")

	distribution := vars["distribution"]

	// execute debian-way/prepare-package.sh from inn-ci-tools
	commands := []string{
		fmt.Sprintf("export MANIFEST_PACKAGE=%s", nameWithVersion),
		fmt.Sprintf("export MANIFEST_INFO_VERSION=%s", version),
		fmt.Sprintf("export PARAM_DISTRIBUTION=%s", distribution),
		fmt.Sprintf("export MANIFEST_INFO_NAME=%s", name),
		fmt.Sprintf("export MANIFEST_BUILD_DEBIAN_DAEMON_ARGS=%s", daemonArgs),
		fmt.Sprintf("export MANIFEST_BUILD_DEBIAN_SERVICE_OWNER=%s", serviceOwner),
		fmt.Sprintf("export MANIFEST_BUILD_DEBIAN_DAEMON_USER=%s", daemonUser),
		fmt.Sprintf("export MANIFEST_BUILD_DEBIAN_DAEMON=%s", daemon),
		fmt.Sprintf("export MANIFEST_BUILD_DEBIAN_DAEMON_PORT=%s", daemonPort),
		fmt.Sprintf("export MANIFEST_BUILD_DEBIAN_MAKE_PIDFILE=%s", makePidfile),
		fmt.Sprintf("export MANIFEST_BUILD_DEBIAN_INSTALL_ROOT=%s", installRoot),
		fmt.Sprintf("export MANIFEST_BUILD_DEBIAN_MAINTAINER_NAME=%s", maintainerName),
		fmt.Sprintf("export MANIFEST_BUILD_DEBIAN_MAINTAINER_EMAIL=%s", maintainerEmail),
		fmt.Sprintf("export MANIFEST_BUILD_DEBIAN_DEPENDS=%s", depends),
		fmt.Sprintf("export MANIFEST_BUILD_DEBIAN_SECTION=%s", section),
		fmt.Sprintf("export MANIFEST_BUILD_DEBIAN_DESCRIPTION=%s", description),
		fmt.Sprintf("export MANIFEST_INFO_CATEGORY=%s", category),
		fmt.Sprintf("export MANIFEST_BUILD_DEBIAN_INIT=%s", init),
		fmt.Sprintf("export MANIFEST_BUILD_DEBIAN_CRON=%s", cron),
		fmt.Sprintf("%s/go/debian-build.sh --distribution=%s", ciToolsPath, distribution),
	}
	for _, cmd := range commands {
		ExecSh(cmd)
		//fmt.Println(color.GreenString(cmd))
	}

	return nil
}

func ExecSh(cmd string) {
	fmt.Println(cmd)
	parts := strings.Fields(cmd)
	out, err := exec.Command(parts[0],parts[1]).Output()
	if err != nil {
		fmt.Println("error occured")
		fmt.Printf("%s", err)
	}
	fmt.Printf("%s", out)
}
