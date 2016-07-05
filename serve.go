package main

import (
	"fmt"
	"log"

	"github.com/fatih/color"
	"gopkg.in/alecthomas/kingpin.v2"

	"github.com/InnovaCo/serve/manifest"

	_ "github.com/InnovaCo/serve/plugins/build"
)

var (
	manifestPath = kingpin.Flag("manifest", "Path to manifest.yml file.").Default("manifest.yml").ExistingFile()
	vars         = *kingpin.Flag("var", "key=value pairs with manifest vars.").StringMap()
	pluginName   = kingpin.Arg("plugin", "Plugin name for run.").Required().String()
)

func main() {
	kingpin.Parse()

	m := manifest.Load(*manifestPath, vars)

	plugins, err := m.FindPlugins(*pluginName)
	if err != nil {
		log.Fatalf("Error find plugins for '%s': %v", *pluginName, err)
	}

	for _, pair := range plugins {
		fmt.Println("")
		log.Println(color.GreenString("%v:", pair.PluginName), pair.Data)

		if err := pair.Plugin.Run(pair.Data); err != nil {
			log.Fatalln("Error on run plugin %s: %v", pair, err)
		}
	}
}
