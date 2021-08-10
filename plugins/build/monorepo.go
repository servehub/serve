package build

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/servehub/utils"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/servehub/serve/manifest"
)

func init() {
	manifest.PluginRegestry.Add("build.monorepo", buildMonorepo{})
}

type buildMonorepo struct{}

func (p buildMonorepo) Run(data manifest.Manifest) error {
	lernaCmd := fmt.Sprintf(data.GetString("command"), data.GetString("lerna-image"))

	log.Println(color.YellowString("> %s", lernaCmd))

	out, err := exec.Command("/bin/bash", "-ec", lernaCmd).CombinedOutput()
	if err != nil {
		log.Println(color.RedString("%s", out))
		return err
	}

	log.Println(color.YellowString("<<< %s", string(out)))

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	pwd, _ := os.Getwd()

	for _, line := range lines {
		if strings.HasPrefix(line, "/src/") {
			dir := pwd + "/" + strings.TrimPrefix(line, "/src/")

			if _, err := os.Stat(dir + "/manifest.yml"); !os.IsNotExist(err) {
				log.Println(color.BlueString("%s was changed", dir))

				if err := utils.RunCmd("cd %s && serve gocd.pipeline.run --branch=%s --commit=%s", dir, data.GetString("branch"), data.GetString("commit")); err != nil {
					log.Println(color.RedString("%s", err))
					return err
				}
			}
		}
	}

	return nil
}
