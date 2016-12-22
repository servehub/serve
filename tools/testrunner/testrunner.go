package testrunner

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"reflect"
	"strings"

	"errors"
	"github.com/codegangsta/cli"
	"github.com/fatih/color"
	"github.com/ghodss/yaml"
)

func TestRunnerCommand() cli.Command {
	return cli.Command{
		Name:  "test-runner",
		Usage: "Run serve with parameters and compare result",
		Flags: []cli.Flag{
			cli.StringFlag{Name: "file", Usage: "Path to yml file with tests"},
			cli.StringFlag{Name: "serve", Value: "serve", Usage: "Path to serve binary file"},
			cli.StringFlag{Name: "config-path", Usage: "Config path with include.d dir, by default = /etc/serve/"},
		},
		Action: func(ctx *cli.Context) error {
			data, err := loadData(ctx.String("file"))
			if err != nil {
				return err
			}

			if _, ok := data["manifest"].(map[string]interface{}); !ok {
				return fmt.Errorf("Key \"manifest\" not found or incorrect type in file %s", ctx.String("file"))
			}

			manifestFile, err := saveManifest(data["manifest"].(map[string]interface{}))
			if err != nil {
				return err
			}

			defer os.Remove(manifestFile) // clean up

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

				testName, ok := testData["name"]
				if !ok {
					testName = testData["run"]
				}

				testData["manifest"] = manifestFile
				if runTest(ctx.String("serve"), ctx.String("config-path"), testData) != nil {
					log.Println(color.RedString("%v: %v ERROR\n", ctx.String("file"), testName))
					counterErr += 1
				} else {
					log.Println(color.GreenString("%v: %v OK\n", ctx.String("file"), testName))
					counterOK += 1
				}
			}

			log.Printf("in total: number of tests %d\n", (counterErr + counterOK))

			if counterErr > 0 {
				return fmt.Errorf(color.RedString("\nTests with errors: %d\n", counterErr))
			}

			return nil
		},
	}
}

func runTest(serve, configPath string, data map[string]interface{}) error {
	run, ok := data["run"]
	if !ok {
		return errors.New("Field `run` is required for test!")
	}

	params := strings.Split(run.(string), " ")
	params = append(params, "--var", fmt.Sprintf("config-path=%v", configPath), "--manifest", data["manifest"].(string), "--dry-run", "--no-color")
	result, err := serveCommand(serve, params...)
	if err != nil {
		return err
	}

	if _, ok := data["expect"].(map[string]interface{}); !ok {
		return fmt.Errorf("Expect is not map type: %v\n", data["expect"])
	}

	if d := diff(result, data["expect"].(map[string]interface{})); !reflect.DeepEqual(d, make(map[string]interface{})) {
		log.Println(color.RedString("diff %v\n", d))
		return fmt.Errorf("Error: diff %v\n", d)
	}
	return nil
}

func serveCommand(serve string, params ...string) (map[string]interface{}, error) {
	print("\n")
	log.Println(color.CyanString("RUN: %v %v\n", serve, strings.Join(params, " ")))

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
			if err := json.Unmarshal(buf.Bytes()[b:e+1], &result); err != nil {
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
		return "", err
	}

	tmpfile, err := ioutil.TempFile("", "serve-test-manifest")
	if err != nil {
		return "", err
	}

	if err := ioutil.WriteFile(tmpfile.Name(), jsonData, 0644); err != nil {
		os.Remove(tmpfile.Name())
		return "", err
	}

	return tmpfile.Name(), nil
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
			switch first[k].(type) {
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
				if first[k] != second[k] {
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
