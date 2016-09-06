package testrunner

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
			cli.StringFlag{Name: "serve", Usage: "serve file"},
			cli.StringFlag{Name: "config-path", Usage: "config path"},
		},
		Action: action,
	}
}

func action(ctx *cli.Context) error {
	data, err := loadData(ctx.String("file"))
	if err != nil {
		return err
	}
	if _, ok := data["manifest"].(map[string]interface{}); !ok {
		return fmt.Errorf("Key \"manifest\" not found or incorrect type in file %s", ctx.String("file"))
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
		if _, ok := test.(map[string]interface{}); !ok {
			log.Printf("Test incorrect format data: %v\n", test)
			continue
		}
		testData := test.(map[string]interface{})
		if _, ok := testData["name"]; !ok {
			log.Printf("Key \"name\" not found in %v\n", testData)
		}
		testData["manifest"] = manifestName
		if runTest(ctx.String("serve"), ctx.String("config-path"), testData) != nil {
			log.Println(color.RedString("%v: %v ERROR\n", ctx.String("file"), testData["name"]))
			counterErr += 1
		} else {
			log.Println(color.GreenString("%v: %v OK\n", ctx.String("file"), testData["name"]))
			counterOK += 1
		}
	}
	log.Printf("in total: number of tests %d\n", (counterErr+counterOK))
	if counterErr > 0 {
		return fmt.Errorf("Tests with errors: %d\n", counterErr)
	}
	return nil
}

func runTest(serve, configPath string, data map[string]interface{}) error {
	params := strings.Split(data["run"].(string), " ")
	params = append(params, "--var", fmt.Sprintf("config-path=%v", configPath), "--manifest", data["manifest"].(string), "--dry-run")
	result, err := serveCommand(serve, params...)
	if err != nil {
		return err
	}
	if _, ok := data["expect"].(map[string]interface{}); !ok {
		return fmt.Errorf("Expect is not map type: %v\n", data["expect"])
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
			result[k] = fmt.Sprintf("%v != <nil>", first[k])
		} else if reflect.TypeOf(first[k]) != reflect.TypeOf(second[k]) {
			result[k] = fmt.Sprintf("type(%v) != type(%v)", first[k], second[k])
		} else {
			switch first[k].(type){
			case map[string]interface{}:
				subResult := diff(first[k].(map[string]interface{}), second[k].(map[string]interface{}))
				if len(subResult) != 0 {
					result[k] = subResult
				}
			case []interface{}:
				if !reflect.DeepEqual(first[k], second[k]) {
					result[k] = fmt.Sprintf("%v != %v", first[k], second[k])
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
			result[k] = fmt.Sprintf("<nil> != %v", second[k])
		}
	}

	return result
}