package consul

import (
	"fmt"
	"log"
	"regexp"

	"github.com/codegangsta/cli"
	"github.com/hashicorp/consul/api"
)

func KvRenameCommand() cli.Command {
	return cli.Command{
		Name:  "rename",
		Usage: "Rename key in consul kv",
		Flags: []cli.Flag{
			cli.StringFlag{Name: "prefix"},
			cli.StringFlag{Name: "from"},
			cli.StringFlag{Name: "to"},
			cli.BoolFlag{Name: "dry-run"},
		},
		Action: func(c *cli.Context) error {
			consul, _ := api.NewClient(api.DefaultConfig())

			list, _, err := consul.KV().List(c.String("prefix"), nil)
			if err != nil {
				return err
			}

			for _, kv := range list {
				from, err := regexp.Compile(c.String("from"))
				if err != nil {
					return fmt.Errorf("Error on parse 'from' regexp: %s", err)
				}

				oldKey := kv.Key
				newKey := from.ReplaceAllString(oldKey, c.String("to"))

				if oldKey != newKey {
					println(kv.Key, "->", newKey)

					if !c.BoolT("dry-run") {
						kv.Key = newKey
						if _, err := consul.KV().Put(kv, nil); err != nil {
							log.Fatalf("Error on rename key '%s': %v", oldKey, err)
						}

						consul.KV().Delete(oldKey, nil)
					}
				}
			}

			return nil
		},
	}
}
