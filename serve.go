package main

import (
	"fmt"
	"log"

	"github.com/fatih/color"
	"gopkg.in/alecthomas/kingpin.v2"

	"github.com/InnovaCo/serve/manifest"
	_ "github.com/InnovaCo/serve/plugins"
)

var version = "0.0"

func main() {
	manifestFile := kingpin.Flag("manifest", "Path to manifest.yml file.").Default("manifest.yml").String()
	plugin := kingpin.Arg("plugin", "Plugin name for run.").String()
	vars := *kingpin.Flag("var", "key=value pairs with manifest vars.").StringMap()
	dryRun := kingpin.Flag("dry-run", "Show manifest section only").Bool()
	noColor := kingpin.Flag("no-color", "Disable colored output").Bool()
	pluginData := kingpin.Flag("plugin-data", "Data for plugin").String()

	kingpin.Version(version)
	kingpin.Parse()

	color.NoColor = *noColor

	var manifestData *manifest.Manifest
	if *pluginData != "" {
		manifestData = manifest.LoadJSON(*pluginData)
	} else {
		manifestData = manifest.Load(*manifestFile, vars)
	}

	if *plugin == "" && *dryRun {
		fmt.Printf("%s\n%s\n%s\n",
			color.GreenString(">>> manifest:"),
			manifestData.String(),
			color.GreenString("<<< manifest: OK\n"))
		return
	}

	var plugins []manifest.PluginData
	if *pluginData != "" {
		plugins = []manifest.PluginData{manifestData.GetPluginWithData(*plugin)}
	} else {
		if result, err := manifestData.FindPlugins(*plugin); err != nil {
			log.Fatalln(color.RedString("Error find plugins for '%s': %v", *plugin, err))
		} else {
			plugins = result
		}
	}

	for _, pair := range plugins {
		log.Printf("%s\n%s\n\n", color.GreenString(">>> %s:", pair.PluginName), color.CyanString("%s", pair.Data))

		if !*dryRun {
			if err := pair.Plugin.Run(pair.Data); err != nil {
				fmt.Println("")
				log.Fatalln(color.RedString("Error on run plugin `%s`: %v", pair.PluginName, err))
			} else {
				log.Println(color.GreenString("<<< %s: OK", pair.PluginName))
			}
		}
	}
}
