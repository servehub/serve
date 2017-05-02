package build

import (
	"github.com/servehub/serve/manifest"
	"github.com/servehub/utils"
)

func init() {
	manifest.PluginRegestry.Add("build.tarball", TarballBuild{})
}

type TarballBuild struct{}

func (p TarballBuild) Run(data manifest.Manifest) error {
	if err := utils.RunCmd("rm -rf ./tarball.tmp && mkdir ./tarball.tmp"); err != nil {
		return err
	}

	for _, f := range data.GetArray("files") {
		if file, ok := f.Unwrap().(string); ok {
			if err := utils.RunCmd("cp -a %s ./tarball.tmp/", file); err != nil {
				return err
			}
		} else if files, ok := f.Unwrap().(map[string]interface{}); ok {
			for from, to := range files {
				if err := utils.RunCmd("cp -a %s ./tarball.tmp/%s", from, to); err != nil {
					return err
				}
			}
		}
	}

	if err := utils.RunCmd("tar -zcf tarball.tar.gz -C ./tarball.tmp/ ."); err != nil {
		return err
	}

	return utils.RunCmd("curl -vsSf -XPUT -T tarball.tar.gz %s", data.GetString("registry-url"))
}
