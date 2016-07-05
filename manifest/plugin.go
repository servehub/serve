package manifest

import "log"

var PluginRegestry = &pluginRegestry{}

type PluginPair struct {
	PluginName string
	Plugin Plugin
	Data Manifest
}

type Plugin interface {
	Run(data Manifest) error
}

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
