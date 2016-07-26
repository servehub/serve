package manifest

import "log"

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

	r.plugins[name] = plugin
}

func (r *pluginRegestry) Get(name string) Plugin {
	p, ok := r.plugins[name]
	if !ok {
		log.Fatalf("Plugin '%s' doesn't exist!", name)
	}
	return p
}
