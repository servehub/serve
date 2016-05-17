package build

import (
	"github.com/InnovaCo/serve/manifest"
	"github.com/InnovaCo/serve/utils"
	"log"
)

type ShellBuild struct{}

func (_ ShellBuild) Run(m *manifest.Manifest, sub *manifest.Manifest) error {
	log.Println("Run shell build", sub)
	return utils.RunCmd(sub.GetString("shell"))
}
