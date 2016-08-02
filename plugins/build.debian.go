package plugins

import (
	"fmt"

	"github.com/InnovaCo/serve/manifest"
	"github.com/InnovaCo/serve/utils"
)

func init() {
	manifest.PluginRegestry.Add("build.debian", BuildDebian{})
}

type BuildDebian struct{}

func (p BuildDebian) Run(data manifest.Manifest) error {
	env := make(map[string]string)

	// required fields
	env["MANIFEST_PACKAGE"] = data.GetString("package")
	env["MANIFEST_INFO_NAME"] = data.GetString("name")
	env["MANIFEST_INFO_VERSION"] = data.GetString("version")
	env["MANIFEST_BUILD_DEBIAN_SECTION"] = data.GetString("category")
	env["MANIFEST_INFO_CATEGORY"] = data.GetString("category")
	env["MANIFEST_BUILD_DEBIAN_MAINTAINER_NAME"] = data.GetString("maintainer-name")
	env["MANIFEST_BUILD_DEBIAN_MAINTAINER_EMAIL"] = data.GetString("maintainer-email")
	env["MANIFEST_BUILD_DEBIAN_INSTALL_ROOT"] = data.GetString("install-root")

	daemon := data.GetString("daemon")
	daemonArgs := data.GetString("daemon-args")

	if daemon != "" {
		daemonArgs = fmt.Sprintf(
			"consul supervisor --service '%s/%s' --port $PORT1 start %s %s",
			data.GetString("category"),
			data.GetString("package"),
			daemon,
			daemonArgs,
		)

		daemon = "/usr/local/bin/serve-tools"
	}

	// optional fields
	env["MANIFEST_BUILD_DEBIAN_DAEMON"] = daemon
	env["MANIFEST_BUILD_DEBIAN_DAEMON_ARGS"] = daemonArgs
	env["MANIFEST_BUILD_DEBIAN_SERVICE_OWNER"] = data.GetString("service-owner")
	env["MANIFEST_BUILD_DEBIAN_DAEMON_USER"] = data.GetString("daemon-user")
	env["MANIFEST_BUILD_DEBIAN_DAEMON_PORT"] = data.GetString("daemon-port")
	env["MANIFEST_BUILD_DEBIAN_MAKE_PIDFILE"] = data.GetString("make-pidfile")
	env["MANIFEST_BUILD_DEBIAN_DEPENDS"] = data.GetString("depends")
	env["MANIFEST_BUILD_DEBIAN_DESCRIPTION"] = data.GetString("description")
	env["MANIFEST_BUILD_DEBIAN_INIT"] = data.GetString("init")
	env["MANIFEST_BUILD_DEBIAN_CRON"] = data.GetString("cron")

	env["GO_PIPELINE_LABEL"] = data.GetString("build-number")
	env["GO_STAGE_COUNTER"] = data.GetString("stage-counter")

	return utils.RunCmdWithEnv(
		"%s/go/debian-build.sh --distribution=%s",
		env,
		data.GetString("ci-tools-path"),
		data.GetString("distribution"),
	)
}
