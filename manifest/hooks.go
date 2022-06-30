package manifest

import (
	"log"

	"github.com/fatih/color"
	"github.com/servehub/utils"
)

type Hooks struct {
	Manifest *Manifest
	DryRun   bool
}

func (h *Hooks) Run(hookName string) error {
	if commands, err := h.Manifest.tree.Search("hooks", hookName).ChildrenMap(); err == nil {
		for k, cmd := range commands {
			log.Println(color.MagentaString("> %s / %s:", hookName, k))

			if !h.DryRun {
				if err := utils.RunCmd(`%s`, cmd.Data()); err != nil {
					log.Println(color.RedString("Error on run %s / %s: %v", hookName, k, err))
					return err
				}
			} else {
				log.Printf("%s", cmd.Data())
			}
		}
	}

	return nil
}
