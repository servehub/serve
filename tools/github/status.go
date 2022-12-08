package github

import (
	"context"
	"fmt"
	"github.com/cenk/backoff"
	"github.com/google/go-github/v44/github"
	"golang.org/x/oauth2"
	"strings"
)

func SendStatus(accessToken string, repo string, ref string, state string, description string, statusContext string, targetUrl string) error {
	if !IsValidState(state) {
		return fmt.Errorf("`%s` is not a valid value for a state", state)
	}

	rp := strings.SplitN(repo, ":", 2)
	rps := strings.SplitN(rp[1], "/", 2)

	input := &github.RepoStatus{
		State:       github.String(state),
		TargetURL:   github.String(targetUrl),
		Description: github.String(description),
		Context:     github.String(statusContext),
	}

	client := github.NewClient(
		oauth2.NewClient(
			context.Background(),
			oauth2.StaticTokenSource(&oauth2.Token{AccessToken: accessToken})))

	return backoff.Retry(func() error {
		_, _, err := client.Repositories.CreateStatus(context.Background(), rps[0], strings.TrimSuffix(rps[1], ".git"), ref, input)
		return err
	}, backoff.WithMaxRetries(backoff.NewExponentialBackOff(), 5))
}

func IsValidState(state string) bool {
	switch state {
	case "error", "failure", "pending", "success":
		return true
	}
	return false
}
