package deploy

import (
	"fmt"
	"strings"

	"github.com/servehub/serve/manifest"
	"github.com/servehub/serve/utils"
)

func init() {
	manifest.PluginRegestry.Add("deploy.tarball", DeployTarball{})
}

type DeployTarball struct{}

func (p DeployTarball) Run(data manifest.Manifest) error {
	if data.GetBool("purge") {
		return p.Uninstall(data)
	} else {
		return p.Install(data)
	}
}

func (p DeployTarball) Install(data manifest.Manifest) error {
	tmp := "tarball-" + utils.RandomString(16)
	dest := data.GetString("install-root") + "/" + data.GetString("package-name")

	hooks := make([]string, 0)
	for _, hook := range data.GetArray("hooks") {
		if hook.Has("postinstall") {
			hooks = append(hooks, "sudo "+hook.GetString("postinstall"))
		}
	}

	hookCmd := strings.Join(hooks, " && ")
	if hookCmd != "" {
		hookCmd = " && " + hookCmd
	}

	/**
	 * todo: register package in consul services catalog
	 */
	if err := utils.RunParallelSshCmd(
		data.GetString("cluster"),
		data.GetString("ssh-user"),
		fmt.Sprint(
			"curl -vsSf -o /tmp/"+tmp+".tar.gz "+data.GetString("package-uri"),
			" && rm -rf /tmp/"+tmp+"/",
			" && mkdir -p /tmp/"+tmp+"/",
			" && tar xzf /tmp/"+tmp+".tar.gz -C /tmp/"+tmp+"/",
			" && sudo rm -rf "+dest,
			" && sudo mkdir -p "+dest,
			" && sudo mv /tmp/"+tmp+"/* "+dest+"/",
			" && sudo chown -R "+data.GetString("user")+":"+data.GetString("group")+" "+dest+"/",
			" && rm -rf /tmp/"+tmp+".tar.gz /tmp/"+tmp+"/",
			hookCmd,
		),
		50,
	); err != nil {
		return err
	}

	return utils.RegisterPluginData("deploy.tarball", data.GetString("package-name"), data.String(), data.GetString("consul-address"))
}

func (p DeployTarball) Uninstall(data manifest.Manifest) error {
	if err := utils.RunParallelSshCmd(
		data.GetString("cluster"),
		data.GetString("ssh-user"),
		fmt.Sprint("sudo rm -rf "+data.GetString("install-root")+"/"+data.GetString("package-name")),
		50,
	); err != nil {
		return err
	}

	return utils.DeletePluginData("deploy.tarball", data.GetString("package-name"), data.GetString("consul-address"))
}
