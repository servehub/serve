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

	log.Println(color.WhiteString("%s", string(out)))

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	pwd, _ := os.Getwd()

	for _, line := range lines {
		if strings.HasPrefix(line, "/src/") {
			dir := pwd + "/" + strings.TrimPrefix(line, "/src/")

			if _, err := os.Stat(dir + "/manifest.yml"); !os.IsNotExist(err) {
				log.Println(color.GreenString("\n :::: %s was changed\n", line))

				if data.GetStringOr("feature", "") == "" {
					if err := utils.RunCmd(`cd %s && serve gocd.pipeline.run --branch="%s" --commit=%s`, dir, data.GetString("branch"), data.GetString("commit")); err != nil {
						log.Println(color.RedString("%s", err))
						return err
					}
				} else {
					for _, task := range data.GetArray("feature-tasks") {
						if err := utils.RunCmd(`cd %s && %s`, dir, task.Unwrap()); err != nil {
							log.Println(color.RedString("%s: %s", task, err))
							return err
						}
					}
				}
			}
		}
	}

	return nil
}
