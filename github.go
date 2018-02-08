package main

import (
	"context"
	"io/ioutil"
	"os"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

const (
	githubAccessTokenKey = "GITHUB_ACCESS_TOKEN"
)

type GitHub struct {
	client      *github.Client
	owner, repo string

	workingDir string
}

func (gh *GitHub) GetFileRaw(path string) (string, error) {
	file, _, _, err := gh.client.Repositories.GetContents(context.Background(), gh.owner, gh.repo, path, nil)
	if err != nil {
		return "", err
	}
	return file.GetContent()
}

func (gh *GitHub) GetFile(path string) (string, error) {
	if content, err := gh.GetFileRaw(path); err != nil {
		return "", err
	} else {
		values, err := ioutil.TempFile(gh.workingDir, "")
		if err != nil {
			return "", err
		}
		if err := ioutil.WriteFile(values.Name(), ([]byte)(content), 0600); err != nil {
			return "", err
		}
		return values.Name(), nil
	}
}

func NewGithubClient(owner, repo, workingDir string) *GitHub {
	token := oauth2.StaticTokenSource(
		&oauth2.Token{
			AccessToken: os.Getenv(githubAccessTokenKey),
		},
	)
	oClient := oauth2.NewClient(context.Background(), token)
	return &GitHub{
		client:     github.NewClient(oClient),
		owner:      owner,
		repo:       repo,
		workingDir: workingDir,
	}
}
