package manifest

import (
	"fmt"
	"log"

	"github.com/fatih/color"
	"github.com/servehub/utils"
)

type Hooks struct {
	Manifest *Manifest
	DryRun   bool
}

func (h *Hooks) Run(hookName string) error {
	if commands, err := h.Manifest.tree.Search("hooks", hookName).Children(); err == nil {
		for _, cmd := range commands {
			log.Println(color.MagentaString("> %s:", hookName))

			if !h.DryRun {
				if err := utils.RunCmd(`%s`, cmd.Data()); err != nil {
					log.Println(color.RedString("Error on run %s: %v", hookName, err))
					return err
				}
			} else {
				log.Printf("%s", cmd.Data())
			}

			fmt.Println("")
		}
	}

	return nil
}
