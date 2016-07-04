package build

import (
	"fmt"

	"github.com/InnovaCo/serve/manifest"
	"github.com/InnovaCo/serve/utils"
)

type MarathonBuild struct{}

func (_ MarathonBuild) Run(m *manifest.Manifest, sub *manifest.Manifest) error {
	if err := utils.RunCmdf("tar -zcf package.tar.gz -C %s/ .", sub.GetString("marathon.package")); err != nil {
		return err
	}

	if err := utils.RunCmdf("curl -vsSf -XPUT -T package.tar.gz %s", TaskRegistryUrl(m)); err != nil {
		return err
	}

	return nil
}

func TaskRegistryUrl(m *manifest.Manifest) string {
	return fmt.Sprintf(
		"http://%s/task-registry/%s/%s-v%s.tar.gz",
		m.GetString("marathon.marathon-host"),
		m.ServiceFullName("/"),
		m.ServiceName(),
		manifest.Escape(m.BuildVersion()),
	)
}
