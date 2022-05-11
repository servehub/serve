package gocd

import (
	"fmt"
	"github.com/servehub/serve/manifest"
)

func init() {
	manifest.PluginRegestry.Add("github.status", githubStatus{})
}

type githubStatus struct{}

func (p githubStatus) Run(data manifest.Manifest) error {

	fmt.Printf("%+v\n", data)

	return nil
}
