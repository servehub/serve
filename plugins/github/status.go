package gocd

import (
	"errors"
	"fmt"
	"github.com/servehub/serve/manifest"
	"github.com/servehub/serve/tools/github"
	"os"
)

func init() {
	manifest.PluginRegestry.Add("github.status", githubStatus{})
}

type githubStatus struct{}

func (p githubStatus) Run(data manifest.Manifest) error {
	accessToken := os.Getenv("GITHUB_TOKEN")
	if accessToken == "" {
		return errors.New("`GITHUB_TOKEN` is required")
	}

	state := data.GetStringOr("state", "success")

	err := github.SendStatus(accessToken,
		data.GetString("repo"),
		data.GetString("ref"),
		state,
		data.GetStringOr("description", fmt.Sprintf("The build status:  %s", state)),
		data.GetStringOr("context", "continuous-integration/serve"),
		data.GetString("target-url"),
	)

	if err != nil {
		fmt.Printf("Github request error: %v\n", err)
	}

	return nil
}
