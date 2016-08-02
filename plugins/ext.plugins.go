package plugins

import (
    "path/filepath"
	"strings"

	"github.com/InnovaCo/serve/manifest"
	"github.com/InnovaCo/serve/utils"
)

func init() {
    if files, err := filepath.Glob("/etc/serve/plugins/*"); err == nil {
        for _, file := range files {
			plugin := ExternalPlugin{Path: file}
            manifest.PluginRegestry.Add(plugin.GetName(), plugin)
        }
    }
}

type ExternalPlugin struct {
    Path string
}

func (p ExternalPlugin)GetName() string {
	s := strings.Split(p.Path, "/")
	return s[len(s)-1:][0]
}

func (p ExternalPlugin) Run(data manifest.Manifest) error {
	return utils.RunCmdWithEnv("%s %s %s", data.ToEnvArray("SERVE_"), p.Path, "--plugin-data", strings.Replace(data.String(), "\n", "", -1))
}
