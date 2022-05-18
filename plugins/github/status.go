package gocd

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/go-github/v44/github"
	"github.com/servehub/serve/manifest"
	"golang.org/x/oauth2"
	"os"
	"strings"
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

	client := github.NewClient(
		oauth2.NewClient(
			context.Background(),
			oauth2.StaticTokenSource(&oauth2.Token{AccessToken: accessToken})))

	rp := strings.SplitN(data.GetString("repo"), ":", 2)
	rps := strings.SplitN(rp[1], "/", 2)

	state := data.GetStringOr("state", "success")
	if !IsValidState(state) {
		return fmt.Errorf("`%s` is not a valid value for a state", state)
	}

	input := &github.RepoStatus{
		State:       github.String(state),
		TargetURL:   github.String(data.GetString("target-url")),
		Description: github.String(data.GetStringOr("description", fmt.Sprintf("The build status:  %s", state))),
		Context:     github.String(data.GetStringOr("context", "continuous-integration/serve")),
	}

	_, _, err := client.Repositories.CreateStatus(context.Background(), rps[0], strings.TrimSuffix(rps[1], ".git"), data.GetString("ref"), input)
	return err
}

func IsValidState(state string) bool {
	switch state {
	case "error", "failure", "pending", "success":
		return true
	}
	return false
}
