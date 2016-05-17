package build

import (
	"github.com/InnovaCo/serve/manifest"
	"github.com/InnovaCo/serve/utils"
)

type ShellBuild struct{}

func (_ ShellBuild) Run(m *manifest.Manifest, sub *manifest.Manifest) error {
	return utils.RunCmd(sub.GetString("shell"))
}
