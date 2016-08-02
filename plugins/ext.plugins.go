package plugins

import (
    "path/filepath"
	"log"
	"os/exec"
	"github.com/fatih/color"
	"strings"
	"github.com/InnovaCo/serve/manifest"
)

func init() {
	log.Println(color.GreenString("external plugins:"))
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
	cmd := exec.Command(p.Path, "--plugin-data", strings.Replace(data.String(), "\n", "", -1))
	cmd.Env = data.ToEnvArray()

	log.Printf("--> %s: ARGS=%s, ENV=%s", p.Path, cmd.Args, cmd.Env)

	out, err := cmd.Output();

	log.Printf("<-- %s: %s", p.Path, string(out))

	return err
}
