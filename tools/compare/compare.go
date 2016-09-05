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
			//cli.StringFlag{Name: "var"},
		},
		Action: serveCommnad,
	}
}


func serveCommnad(ctx *cli.Context) error {
	//fmt.Printf("--> %s %s %s %s %s\n", ctx.String("serve"), ctx.String("plugin"), ctx.String("manifest"), ctx.String("result"), ctx.String("var"))
	resultFile, err := os.Open(ctx.String("result"))
	defer resultFile.Close()
	if err != nil {
		log.Printf("Cannot open %s", ctx.String("result"))
		return fmt.Errorf("Cannot open %s", ctx.String("result"))
	}

	resBytes, err := ioutil.ReadAll(resultFile)
	if err != nil {
		log.Printf("Cannot open %s", ctx.String("result"))
		return fmt.Errorf("Cannot open %s", ctx.String("result"))
	}

	resBytes, err = yaml.YAMLToJSON(resBytes)
	if err != nil {
		log.Printf("Error on parse %s: %v!", ctx.String("result"), err)
		return fmt.Errorf("Error on parse %s: %v!", ctx.String("result"), err)
	}

	var resJSON map[string]interface{}
	err = json.Unmarshal(resBytes, &resJSON)
	if err != nil {
		log.Printf("Cannot parse %s: %v", ctx.String("result"), err)
		return fmt.Errorf("Cannot parse %s: %v", ctx.String("result"), err)
	}

	//fmt.Printf("\n%v\n", string(resBytes))
	//vars := ""

	//if ctx.String("var") != "" {
	//
	//}
	cmd := exec.Command(ctx.String("serve")+"/serve", ctx.String("plugin"), "--manifest", ctx.String("manifest"), "--dry-run")

	cmd.Env = os.Environ()

	var buf bytes.Buffer
	cmd.Stderr = &buf

	if err := cmd.Run(); err != nil {
		log.Printf("%v\n", err)
		return err
	}

	var inJSON map[string]interface{}

	if b := strings.Index(buf.String(), "{"); b != -1{
		if e := strings.Index(buf.String(), "}\n\n"); e != -1 {
			if err := json.Unmarshal(buf.Bytes()[b:e + 1], &inJSON); err != nil {
				log.Printf("%v\n", err)
				return err
			}
			//fmt.Printf("\n%s\n", string(buf.Bytes()[b:e + 1]))
		}
	}

	if d := diff(inJSON, resJSON); !reflect.DeepEqual(d, make(map[string]interface{})) {
		log.Printf("diff %v\n", d)
		return fmt.Errorf("diff %v\n", d)
	}

	log.Println("Ok")

	return nil
}

func diff(first map[string]interface{}, second map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for k := range first {
		if _, ok := second[k]; !ok {
			result[k] = "undefined"
		} else if reflect.TypeOf(first[k]) != reflect.TypeOf(second[k]) {
			result[k] = second[k]
		} else {
			switch first[k].(type){
			default:
				if(first[k] != second[k]) {
					result[k] = second[k]
				}
			case map[string]interface{}:
				subResult := diff(first[k].(map[string]interface{}), second[k].(map[string]interface{}))
				if len(subResult) != 0 {
					result[k] = subResult
				}
			case []interface{}:
				if !reflect.DeepEqual(first[k], second[k]) {
					result[k] = second[k]
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