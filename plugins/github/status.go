package gocd

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/go-github/v44/github"
	"github.com/servehub/serve/manifest"
	"golang.org/x/oauth2"
	"os"
)

func IsValidState(state string) bool {
	switch state {
	case "error", "failure", "pending", "success":
		return true
	}
	return false
}

func init() {
	manifest.PluginRegestry.Add("github.status", githubStatus{})
}

type githubStatus struct{}

func (p githubStatus) Run(data manifest.Manifest) error {
	accessToken := os.Getenv("GITHUB_API_OAUTH_TOKEN")
	if accessToken == "" {
		return errors.New("`GITHUB_API_OAUTH_TOKEN` is required")
	}

	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: accessToken})
	ctx := context.Background()
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	owner := data.GetString("owner")
	repo := data.GetString("repo")
	ref := data.GetString("ref")
	if owner == "" || repo == "" || ref == "" {
		return errors.New("`owner`, `repo` and `ref` are required options")
	}
	state := data.GetStringOr("state", "success")
	if !IsValidState(state) {
		return fmt.Errorf("`%s` is not a valid value for a state", state)
	}
	targetUrl := data.GetString("target-url")
	description := data.GetStringOr("description", fmt.Sprintf("The build status:  %s", state))
	contextValue := data.GetStringOr("context", "continuous-integration/serve")
	input := &github.RepoStatus{
		State:       github.String(state),
		TargetURL:   github.String(targetUrl),
		Description: github.String(description),
		Context:     github.String(contextValue),
	}
	_, _, err := client.Repositories.CreateStatus(ctx, owner, repo, ref, input)
	if err != nil {
		return err
	}
	return nil
}
