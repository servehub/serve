package build

import (
	"github.com/InnovaCo/serve/manifest"
	"github.com/InnovaCo/serve/utils"
)

func init() {
	manifest.PluginRegestry.Add("build.sbt", SbtBuild{})
}

type SbtBuild struct{}

func (p SbtBuild) Run(data manifest.Manifest) error {
  return utils.RunCmd(
	  `sbt ';set every version := "%s"' clean "%s" %s`,
	  data.GetString("version"),
	  data.GetString("test"),
	  data.GetStringOr("sbt", ""),
  )
}
