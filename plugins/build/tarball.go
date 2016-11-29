package build

import (
	"github.com/InnovaCo/serve/manifest"
	"github.com/InnovaCo/serve/utils"
)

func init() {
	manifest.PluginRegestry.Add("build.tarball", TarballBuild{})
}

type TarballBuild struct{}

func (p TarballBuild) Run(data manifest.Manifest) error {
	utils.RunCmd("rm -rf ./tarball.tmp && mkdir ./tarball.tmp")

	for _, f := range data.GetArray("files") {
		if file, ok := f.Unwrap().(string); ok {
			utils.RunCmd("cp -a %s ./tarball.tmp/", file)
		} else if files, ok := f.Unwrap().(map[string]interface{}); ok {
			for from, to := range files {
				utils.RunCmd("cp -a %s ./tarball.tmp/%s", from, to)
			}
		}
	}

	if err := utils.RunCmd("tar -zcf tarball.tar.gz -C ./tarball.tmp/ ."); err != nil {
		return err
	}

	return utils.RunCmd("curl -vsSf -XPUT -T tarball.tar.gz %s", data.GetString("registry-url"))
}
