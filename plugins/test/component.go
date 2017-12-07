package test

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/fatih/color"
	"github.com/ghodss/yaml"
	"github.com/servehub/serve/manifest"
	"github.com/servehub/utils"
)

func init() {
	manifest.PluginRegestry.Add("test.component", TestComponent{})
}

type TestComponent struct{}

func (p TestComponent) Run(data manifest.Manifest) error {
	if data.GetString("env") != data.GetString("current-env") {
		log.Printf("No component test found for `%s`.\n", data.GetString("current-env"))
		return nil
	}

	bytes, err := yaml.Marshal(data.GetTree("compose").Unwrap())
	if err != nil {
		return fmt.Errorf("error on serialize yaml: %v", err)
	}

	tmpfile, err := ioutil.TempFile("", "serve-component-test")
	if err != nil {
		return fmt.Errorf("error create tmpfile: %v", err)
	}

	defer func() {
		tmpfile.Close()
		os.Remove(tmpfile.Name())
	}()

	if _, err := tmpfile.Write(bytes); err != nil {
		return fmt.Errorf("error write to tmpfile: %v", err)
	}

	if err := utils.RunCmd("docker-compose -p %s -f %s pull --parallel", data.GetString("name"), tmpfile.Name()); err != nil {
		return fmt.Errorf("error on pull new images for docker-compose: %v", err)
	}

	defer func() {
		utils.RunCmd("docker-compose -p %s -f %s down -v --remove-orphans", data.GetString("name"), tmpfile.Name())
	}()

	go func() {
		select {
		case <-time.After(5 * time.Minute):
			color.Red("Timeout exceeded for tests, exit...")
			utils.RunCmd("docker-compose -p %s -f %s down -v --remove-orphans", data.GetString("name"), tmpfile.Name())
		}
	}()

	return utils.RunCmd("docker-compose -p %s -f %s up --remove-orphans --force-recreate --abort-on-container-exit", data.GetString("name"), tmpfile.Name())
}
