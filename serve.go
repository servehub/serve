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
	dryRun 	 	 := kingpin.Flag("dry-run", "Show manifest section only").Bool()
	pluginData 	 := kingpin.Flag("plugin-data", "Data for plugin").String()

	kingpin.Version(version)
	kingpin.Parse()

	var mnf *manifest.Manifest
	var plugins []manifest.PluginPair
	var err error
	if pluginData == nil {
		mnf = manifest.Load(*manifestFile, vars)
		plugins, err = mnf.FindPlugins(*plugin)
	} else {
		fmt.Println(*pluginData)
		mnf = manifest.LoadJSON(*pluginData)
		plugins = append(plugins, mnf.GetPluginWithData(*plugin))
	}

	if err != nil {
		log.Fatalf("Error find plugins for '%s': %v", *plugin, err)
	}

	for _, pair := range plugins {
		log.Println(color.GreenString("%v:\n", pair.PluginName), pair.Data)

		if !*dryRun {
			if err := pair.Plugin.Run(pair.Data); err != nil {
				fmt.Println("")
				log.Fatalln(color.RedString("Error on run plugin `%s`: %v", pair.PluginName, err))
			}
		}
	}
}
