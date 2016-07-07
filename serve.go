package main

import (
	"fmt"
	"log"

	"github.com/fatih/color"
	"gopkg.in/alecthomas/kingpin.v2"

	"github.com/InnovaCo/serve/manifest"
	_ "github.com/InnovaCo/serve/plugins"
)

var (
	manifestPath = kingpin.Flag("manifest", "Path to manifest.yml file.").Default("manifest.yml").String()
	plugin       = kingpin.Arg("plugin", "Plugin name for run.").Required().String()
	vars         = *kingpin.Flag("var", "key=value pairs with manifest vars.").StringMap()
)

func main() {
	kingpin.Parse()

	m := manifest.Load(*manifestPath, vars)

	plugins, err := m.FindPlugins(*plugin)
	if err != nil {
		log.Fatalf("Error find plugins for '%s': %v", *plugin, err)
	}

	for _, pair := range plugins {
		fmt.Println("")
		log.Println(color.GreenString("%v:", pair.PluginName), pair.Data)

		if err := pair.Plugin.Run(pair.Data); err != nil {
			log.Fatalln("Error on run plugin %s: %v", pair, err)
		}
	}
}
