package compare

import (
	"fmt"
	"os"
	"encoding/json"
	"reflect"
	"os/exec"
	"strings"
	"bytes"
	"io/ioutil"
	"log"

	"github.com/ghodss/yaml"
	"github.com/codegangsta/cli"
)

func CompareCommand() cli.Command {
	return cli.Command{
		Name:  "compare",
		Usage: "Run serve with parameters and compare result",
		Flags: []cli.Flag{
			cli.StringFlag{Name: "serve", Usage: "serve dir path"},
			cli.StringFlag{Name: "plugin", Usage: "plugin name"},
			cli.StringFlag{Name: "manifest", Usage: "manifest file name"},
			cli.StringFlag{Name: "result", Usage: "result file name"},
		},
		Action: action,
	}
}

func serveCommand(serveDir, plugin, manifest string) (map[string]interface{}, error) {
	cmd := exec.Command(serveDir+"/serve", plugin, "--manifest", manifest, "--dry-run")
	cmd.Env = os.Environ()
	buf := bytes.Buffer{}
	cmd.Stderr = &buf

	if err := cmd.Run(); err != nil {
		log.Printf("%v\n", err)
		return nil, err
	}
	result := make(map[string]interface{})
	if b := strings.Index(buf.String(), "{"); b != -1 {
		if e := strings.Index(buf.String(), "}\n\n"); e != -1 {
			if err := json.Unmarshal(buf.Bytes()[b:e + 1], &result); err != nil {
				log.Printf("%v\n", err)
				return nil, err
			}
		}
	}
	return result, nil
}

func loadResult(name string) (map[string]interface{}, error) {
	resultFile, err := os.Open(name)
	defer resultFile.Close()
	if err != nil {
		log.Printf("Cannot open %s", name)
		return nil, fmt.Errorf("Cannot open %s", name)
	}
	resBytes, err := ioutil.ReadAll(resultFile)
	if err != nil {
		log.Printf("Cannot open %s", name)
		return nil, fmt.Errorf("Cannot open %s", name)
	}
	result := make(map[string]interface{})
	if err := yaml.Unmarshal(resBytes, &result); err != nil {
		log.Printf("Cannot parse %s: %v", name, err)
		return nil, fmt.Errorf("Cannot parse %s: %v", name, err)
	}
	return result, nil
}

func action(ctx *cli.Context) error {
	resJSON, err := loadResult(ctx.String("result"))
	if err != nil {
		log.Printf("%v\n", err)
		return err
	}
	inJSON, err := serveCommand(ctx.String("serve"), ctx.String("plugin"), ctx.String("manifest"))
	if err != nil {
		log.Printf("%v\n", err)
		return err
	}
	if d := diff(inJSON, resJSON); !reflect.DeepEqual(d, make(map[string]interface{})) {
		log.Printf("diff %v\n", d)
		return fmt.Errorf("Error diff: %v\n", d)
	}
	log.Println("Ok")
	return nil
}

func diff(first map[string]interface{}, second map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for k := range first {
		if _, ok := second[k]; !ok {
			result[k] = second[k]
		} else if reflect.TypeOf(first[k]) != reflect.TypeOf(second[k]) {
			result[k] = second[k]
		} else {
			switch first[k].(type){
			case map[string]interface{}:
				subResult := diff(first[k].(map[string]interface{}), second[k].(map[string]interface{}))
				if len(subResult) != 0 {
					result[k] = subResult
				}
			case []interface{}:
				if !reflect.DeepEqual(first[k], second[k]) {
					result[k] = second[k]
				}
			default:
				if(first[k] != second[k]) {
					result[k] = fmt.Sprintf("%v != %v", first[k], second[k])
				}
			}
		}
	}
	for k := range second {
		if _, ok := first[k]; !ok {
			result[k] = second[k]
		}
	}

	return result
}