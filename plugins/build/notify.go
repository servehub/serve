package build

import (
	"fmt"
	"strings"

	"github.com/servehub/serve/manifest"
	"github.com/servehub/utils"
)

func init() {
	manifest.PluginRegestry.Add("build.notify", BuildNotify{})
}

type BuildNotify struct{}

func (p BuildNotify) Run(data manifest.Manifest) error {
	manifestArg := ""
	if data.Has("manifest") {
		manifestArg = fmt.Sprintf(`--manifest="%s"`, data.GetString("manifest"))
	}

	return utils.RunCmd(
		`serve notify --env=%s --event="%s" --message="%s" --build-number="%s" --changelog-for="%s" %s`,
		data.GetString("env"),
		data.GetString("event"),
		strings.Replace(data.GetString("message"), `"`, `”`, -1),
		data.GetString("build-number"),
		data.GetString("changelog-for"),
		manifestArg,
	)
}
