package main

import (
	"fmt"
	"log"

	"github.com/fatih/color"
	"gopkg.in/alecthomas/kingpin.v2"

	"github.com/InnovaCo/serve/manifest"
	_ "github.com/InnovaCo/serve/plugins"
)

var version = "1.3"

func init() {
	color.NoColor = false
}

func main() {
	manifestFile := kingpin.Flag("manifest", "Path to manifest.yml file.").Default("manifest.yml").String()
	plugin       := kingpin.Arg("plugin", "Plugin name for run.").String()
	vars         := *kingpin.Flag("var", "key=value pairs with manifest vars.").StringMap()
	dryRun       := kingpin.Flag("dry-run", "Show manifest section only").Bool()
	pluginData   := kingpin.Flag("plugin-data", "Data for plugin").String()

	kingpin.Version(version)
	kingpin.Parse()

	var plugins []manifest.PluginData
	var err error
	var manifestData *manifest.Manifest

	if *pluginData != "" {
		manifestData = manifest.LoadJSON(*pluginData)
	} else {
		manifestData = manifest.Load(*manifestFile, vars)
	}

	if *plugin == "" && *dryRun {
		fmt.Printf("%s\n%s\n%s\n", color.GreenString(">>> manifest:"),
			                       manifestData.String(),
			                       color.GreenString("<<< manifest: OK\n"))
		return
	}

	if *pluginData != "" {
		plugins = []manifest.PluginData{manifestData.GetPluginWithData(*plugin)}
	} else {
		plugins, err = manifestData.FindPlugins(*plugin)
	}

	if err != nil {
		log.Fatalln(color.RedString("Error find plugins for '%s': %v", *plugin, err))
	}

	for _, pair := range plugins {
		log.Printf("%s\n%s\n\n", color.GreenString(">>> %s:", pair.PluginName), pair.Data)

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
