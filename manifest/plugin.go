package manifest

import (
	"log"
	"github.com/fatih/color"
)

type Plugin interface {
	Run(data Manifest) error
}

type PluginData struct {
	PluginName string
	Plugin Plugin
	Data Manifest
}

var PluginRegestry = &pluginRegestry{}

type pluginRegestry struct {
	plugins map[string]Plugin
}

func (r *pluginRegestry) Add(name string, plugin Plugin) {
	if r.plugins == nil {
		r.plugins = make(map[string]Plugin)
	}

	if _, ok := r.plugins[name]; ok {
		log.Fatalf(color.RedString("Plugin '%s' dublicate name", name))
	}

	log.Printf(color.GreenString("add plugin: %s", name))

	r.plugins[name] = plugin
}

func (r *pluginRegestry) Get(name string) Plugin {
	p, ok := r.plugins[name]
	if !ok {
		log.Fatalf(color.RedString("Plugin '%s' doesn't exist!", name))
	}
	return p
}
