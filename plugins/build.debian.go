package plugins

import (
	"github.com/InnovaCo/serve/manifest"
	"fmt"
	"github.com/fatih/color"
	"os"
	"log"
	"github.com/InnovaCo/serve/utils"
)

const ciToolsPath = "/var/go/inn-ci-tools"

func init() {
	manifest.PluginRegestry.Add("build.debian", BuildDebian{})
}

type BuildDebian struct {}

func(p BuildDebian) Run(data manifest.Manifest, vars map[string]string) error {
	log.Println(color.GreenString("Start build.debian plugin"))
	var exports map[string]string = make(map[string]string)

	nameVersion, name, version, category, installRoot, maintainerName, maintainerEmail,
	section, daemonArgs, serviceOwner, daemonUser, daemon, daemonPort, makePidfile,
	depends, description, init, cron :=
		data.GetString("name-version"),
		data.GetString("name"),
		data.GetString("version"),
		data.GetString("category"),
		data.GetString("install-root"),
		data.GetString("maintainer-name"),
		data.GetString("maintainer-email"),
		data.GetString("section"),
		data.GetString("daemon-args"),
		data.GetString("service-owner"),
		data.GetString("daemon-user"),
		data.GetString("daemon"),
		data.GetString("daemon-port"),
		data.GetString("make-pidfile"),
		data.GetString("depends"),
		data.GetString("description"),
		data.GetString("init"),
		data.GetString("cron")

	// required fields
	exports["MANIFEST_PACKAGE"] = nameVersion
	exports["MANIFEST_INFO_NAME"] = name
	exports["MANIFEST_INFO_VERSION"] = version
	exports["MANIFEST_BUILD_DEBIAN_SECTION"] = section
	exports["MANIFEST_INFO_CATEGORY"] = category
	exports["MANIFEST_BUILD_DEBIAN_MAINTAINER_NAME"] = maintainerName
	exports["MANIFEST_BUILD_DEBIAN_MAINTAINER_EMAIL"] = maintainerEmail
	exports["MANIFEST_BUILD_DEBIAN_INSTALL_ROOT"] = installRoot
	// optional fields
	exports["MANIFEST_BUILD_DEBIAN_DAEMON_ARGS"] = daemonArgs
	exports["MANIFEST_BUILD_DEBIAN_SERVICE_OWNER"] = serviceOwner
	exports["MANIFEST_BUILD_DEBIAN_DAEMON_USER"] = daemonUser
	exports["MANIFEST_BUILD_DEBIAN_DAEMON"] = daemon
	exports["MANIFEST_BUILD_DEBIAN_DAEMON_PORT"] = daemonPort
	exports["MANIFEST_BUILD_DEBIAN_MAKE_PIDFILE"] = makePidfile
	exports["MANIFEST_BUILD_DEBIAN_DEPENDS"] = depends
	exports["MANIFEST_BUILD_DEBIAN_DESCRIPTION"] = description
	exports["MANIFEST_BUILD_DEBIAN_INIT"] = init
	exports["MANIFEST_BUILD_DEBIAN_CRON"] = cron

	distribution := vars["distribution"]

	fmt.Println(color.GreenString("Start exporting vars"))
	for key, val := range exports {
		fmt.Println(color.GreenString("export %s=%s", key, val))
		os.Setenv(key, val)
	}
	// call debian-build.sh from inn-ci-tools
	return utils.RunCmd(
		fmt.Sprintf("%s/go/debian-build.sh --distribution=%s", ciToolsPath, distribution))
}
