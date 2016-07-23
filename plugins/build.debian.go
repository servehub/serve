package plugins

import (
	"github.com/InnovaCo/serve/manifest"
	"fmt"
	"github.com/fatih/color"
	"os"
	"log"
	"github.com/InnovaCo/serve/utils"
)

func init() {
	manifest.PluginRegestry.Add("build.debian", BuildDebian{})
}

type BuildDebian struct {}

func(p BuildDebian) Run(data manifest.Manifest) error {
	var exports map[string]string = make(map[string]string)
	// required fields
	category := data.GetString("category")

	exports["MANIFEST_PACKAGE"] = data.GetString("name-version")
	exports["MANIFEST_INFO_NAME"] = data.GetString("name")
	exports["MANIFEST_INFO_VERSION"] = data.GetString("version")
	exports["MANIFEST_BUILD_DEBIAN_SECTION"] = category
	exports["MANIFEST_INFO_CATEGORY"] = category
	exports["MANIFEST_BUILD_DEBIAN_MAINTAINER_NAME"] = data.GetString("maintainer-name")
	exports["MANIFEST_BUILD_DEBIAN_MAINTAINER_EMAIL"] = data.GetString("maintainer-email")
	exports["MANIFEST_BUILD_DEBIAN_INSTALL_ROOT"] = data.GetString("install-root")
	// optional fields
	exports["MANIFEST_BUILD_DEBIAN_DAEMON_ARGS"] = data.GetString("daemon-args")
	exports["MANIFEST_BUILD_DEBIAN_SERVICE_OWNER"] = data.GetString("service-owner")
	exports["MANIFEST_BUILD_DEBIAN_DAEMON_USER"] = data.GetString("daemon-user")
	exports["MANIFEST_BUILD_DEBIAN_DAEMON"] = data.GetString("daemon")
	exports["MANIFEST_BUILD_DEBIAN_DAEMON_PORT"] = data.GetString("daemon-port")
	exports["MANIFEST_BUILD_DEBIAN_MAKE_PIDFILE"] = data.GetString("make-pidfile")
	exports["MANIFEST_BUILD_DEBIAN_DEPENDS"] = data.GetString("depends")
	exports["MANIFEST_BUILD_DEBIAN_DESCRIPTION"] = data.GetString("description")
	exports["MANIFEST_BUILD_DEBIAN_INIT"] = data.GetString("init")
	exports["MANIFEST_BUILD_DEBIAN_CRON"] = data.GetString("cron")
	exports["GO_STAGE_COUNTER"] = data.GetString("stage-counter")

	distribution := data.GetString("distribution")
	ciToolsPath := data.GetString("ci-tools-path")

	log.Println(color.GreenString("Start exporting vars"))
	for key, val := range exports {
		fmt.Println(color.GreenString("export %s=%s", key, val))
		os.Setenv(key, val)
	}
	// call debian-build.sh from inn-ci-tools
	return utils.RunCmd(
		fmt.Sprintf("%s/go/debian-build.sh --distribution=%s", ciToolsPath, distribution))
}
