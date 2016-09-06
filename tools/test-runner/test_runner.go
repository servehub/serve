package test_runner

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

	"github.com/fatih/color"

	"github.com/ghodss/yaml"
	"github.com/codegangsta/cli"
)

func TestRunnerCommand() cli.Command {
	return cli.Command{
		Name:  "test-runner",
		Usage: "Run serve with parameters and compare result",
		Flags: []cli.Flag{
			cli.StringFlag{Name: "file", Usage: "file name with tests"},
			cli.StringFlag{Name: "serve", Usage: "serve path"},
		},
		Action: action,
	}
}

func action(ctx *cli.Context) error {
	data, err := loadData(ctx.String("file"))
	if err != nil {
		return err
	}
	if _, ok := data["manifest"]; !ok {
		return fmt.Errorf("Key \"manifest\" not found in file %s", ctx.String("file"))
	}
	manifestName, err := saveManifest(data["manifest"].(map[string]interface{}))
	if err != nil {
		return err
	}

	if _, ok := data["tests"]; !ok {
		return fmt.Errorf("Key \"tests\" not found in file %s", ctx.String("file"))
	}

	counterOK := 0
	counterErr := 0
	for _, test := range data["tests"].([]interface{}) {
		test.(map[string]interface{})["manifest"] = manifestName
		if runTest(ctx.String("serve"), test.(map[string]interface{})) != nil {
			log.Println(color.RedString("%v: %v ERROR\n", ctx.String("file"), test.(map[string]interface{})["name"]))
			counterErr += 1
		} else {
			log.Println(color.GreenString("%v: %v OK\n", ctx.String("file"), test.(map[string]interface{})["name"]))
			counterOK += 1
		}
	}

	log.Printf("Stat: All test %d\n", (counterErr+counterOK))
	if counterErr > 0 {
		return fmt.Errorf("Tests with errors: %d\n", counterErr)
	}
	return nil
}

func runTest(serve string, data map[string]interface{}) error {
	params := strings.Split(data["params"].(string), " ")
	params = append(params, "--manifest", data["manifest"].(string), "--dry-run")
	result, err := serveCommand(serve, params...)
	if err != nil {
		return err
	}
	if d := diff(result, data["expect"].(map[string]interface{})); !reflect.DeepEqual(d, make(map[string]interface{})) {
		log.Printf("diff %v\n", d)
		return fmt.Errorf("Error diff: %v\n", d)
	}
	return nil
}

func serveCommand(serve string, params...string) (map[string]interface{}, error) {
	log.Printf("RUN: %v %v\n", serve, params)

	cmd := exec.Command(serve, params...)
	cmd.Env = os.Environ()
	buf := bytes.Buffer{}
	cmd.Stderr = &buf

	if err := cmd.Run(); err != nil {
		log.Printf("serve error: %v\n%v\n", err, buf.String())
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

func saveManifest(data map[string]interface{}) (string, error) {
	jsonData, err := yaml.Marshal(&data)
	if err != nil {
		log.Fatalf("error: %v", err)
		return "", err
	}
    if err := ioutil.WriteFile("/tmp/serve_test_manifest.yml", jsonData, 0644); err != nil {
		return "", err
	}
	return "/tmp/serve_test_manifest.yml", nil
}

func loadData(name string) (map[string]interface{}, error) {
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