package consul

import (
	"encoding/json"

	"github.com/codegangsta/cli"
	"github.com/hashicorp/consul/api"

	"github.com/InnovaCo/serve/utils/mergemap"
)

func KvPatchCommand() cli.Command {
	return cli.Command{
		Name:  "kv-patch",
		Usage: "Patch json values in consul kv",
		Flags: []cli.Flag{
			cli.StringFlag{Name: "key"},
			cli.StringFlag{Name: "value"},
			cli.BoolFlag{Name: "dry-run"},
		},
		Action: func(c *cli.Context) error {
			consul, _ := api.NewClient(api.DefaultConfig())

			list, _, err := consul.KV().List(c.String("key"), nil)
			if err != nil {
				return err
			}

			for _, kv := range list {
				consulValue := make(map[string]interface{})
				if err := json.Unmarshal(kv.Value, &consulValue); err != nil {
					return err
				}

				patchValue := make(map[string]interface{})
				if err := json.Unmarshal([]byte(c.String("value")), &patchValue); err != nil {
					return err
				}

				merged, err := mergemap.Merge(consulValue, patchValue)
				if err != nil {
					return err
				}

				mergedJson, _ := json.MarshalIndent(merged, "", "  ")
				println(kv.Key)
				println(string(mergedJson) + "\n")

				if c.Bool("dry-run") {
					continue
				}

				_, putErr := consul.KV().Put(&api.KVPair{Key: kv.Key, Value: mergedJson}, nil)
				if putErr != nil {
					return putErr
				}
			}

			return nil
		},
	}
}
