package build

import (
	"github.com/InnovaCo/serve/manifest"
	"github.com/InnovaCo/serve/utils"
)

type MarathonBuild struct{}

func (_ MarathonBuild) Run(m *manifest.Manifest, sub *manifest.Manifest) error {
	if err := utils.RunCmdf("tar -zcf package.tar.gz -C %s/ .", sub.GetString("marathon.package")); err != nil {
		return err
	}

	if err := utils.RunCmdf(
		"curl -vsSf -XPUT -T package.tar.gz http://%s/task-registry/%s/%s-%s.tar.gz",
		sub.GetString("marathon.marathon-host"),
		m.ServiceName(),
		m.ServiceName(),
		m.BuildVersion()); err != nil {
		return err
	}

	return nil
}
