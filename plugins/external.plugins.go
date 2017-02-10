package plugins

import (
	"fmt"
	"log"
	"path/filepath"
	"strings"

	"github.com/servehub/serve/manifest"
	"github.com/servehub/serve/utils"
)

func init() {
	files, err := filepath.Glob("/etc/serve/plugins/*")
	if err != nil {
		log.Fatalln("Error on load external plugins from /etc/serve/plugins/*", err)
	}

	for _, file := range files {
		plugin := ExternalPlugin{Path: file}
		manifest.PluginRegestry.Add(plugin.Name(), plugin)
	}
}

type ExternalPlugin struct {
	Path string
}

func (p ExternalPlugin) Name() string {
	return p.Path[strings.LastIndex(p.Path, "/")+1:]
}

func (p ExternalPlugin) Run(data manifest.Manifest) error {
	return utils.RunCmdWithEnv(fmt.Sprintf("%s --plugin-data '%s'", p.Path, strings.Replace(data.String(), "\n", "", -1)), data.ToEnvMap("SERVE_"))
}
