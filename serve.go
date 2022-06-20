package main

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"time"

	"github.com/fatih/color"

	"github.com/servehub/serve/manifest"
	_ "github.com/servehub/serve/plugins"
)

var version = "0.0"

var flagRegex = regexp.MustCompile("--([a-z0-9-]+)(=(.+))?")
var pluginNameRegex = regexp.MustCompile("^[a-z][a-z0-9-.]+$")

func main() {
	plugin := ""
	if len(os.Args) > 1 && pluginNameRegex.MatchString(os.Args[1]) {
		plugin = os.Args[1]
	}

	vars := make(map[string]string, 0)
	for _, arg := range os.Args[1:] {
		res := flagRegex.FindStringSubmatch(arg)

		if len(res) > 0 {
			if res[2] == "" {
				vars[res[1]] = "true"
			} else {
				vars[res[1]] = res[3]
			}
		}
	}

	color.NoColor = false
	if _, ok := vars["no-color"]; ok {
		color.NoColor = true
	}

	manifestFile := "manifest.yml"
	if f, ok := vars["manifest"]; ok {
		manifestFile = f
	}

	if _, ok := vars["version"]; ok && plugin == "" {
		fmt.Printf("v%s\n", version)
		return
	}

	pluginDataFile, pluginDataExists := vars["plugin-data"]

	var manifestData *manifest.Manifest
	if pluginDataExists {
		manifestData = manifest.LoadJSON(pluginDataFile)
	} else {
		manifestData = manifest.Load(manifestFile, plugin, vars)
	}

	_, dryRun := vars["dry-run"]

	if dryRun && plugin == "" {
		fmt.Printf("%s\n%s\n%s\n", color.GreenString(">>> manifest:"), manifestData.String(), color.GreenString("<<< manifest: OK\n"))
		return
	}

	pluginFilter, pluginFilterExists := vars["plugin"]

	var plugins []manifest.PluginData
	if pluginDataExists {
		plugins = []manifest.PluginData{manifestData.GetPluginWithData(plugin)}
	} else {
		if result, err := manifestData.FindPlugins(plugin); err != nil {
			log.Fatalln(color.RedString("Error find plugins for '%s': %v", plugin, err))
		} else {
			plugins = result
		}
	}

	startTime := time.Now()

	for _, pair := range plugins {
		if pluginFilterExists && pair.PluginName != pluginFilter {
			continue
		}

		log.Printf("%s\n%s\n\n", color.GreenString(">>> %s:", pair.PluginName), color.CyanString("%s", pair.Data))

		if !dryRun {
			if err := pair.Plugin.Run(pair.Data); err != nil {
				fmt.Println("")
				log.Fatalln(color.RedString("Error on run plugin `%s`: %v", pair.PluginName, err))
			} else {
				log.Println(color.GreenString("<<< %s: OK", pair.PluginName))
			}
		}
	}

	log.Println(color.GreenString("Time: %d seconds", int(time.Now().Sub(startTime).Seconds())))
}
