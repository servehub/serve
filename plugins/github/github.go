package github

type Push struct {
	Ref        string         `json:"ref"`
	Repository GithubRepo     `json:"repository"`
	Commits    []GithubCommit `json:"commits"`
}

type GithubRepo struct {
	SshUrl     string `json:"ssh_url"`
	ContentUrl string `json:"contents_url"`
}

type GithubCommit struct {
	Modified []string `json:"modified"`
	Added    []string `json:"added"`
	Removed  []string `json:"removed"`
}

type FileContent struct {
	Sha     string `json:"sha"`
	Content string `json:"content"`
}
