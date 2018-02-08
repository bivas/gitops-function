package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

const (
	kubeconfigRepoUrlKey = "KUBECONFIG_REPO_URL"
)

func getKubeConfig(tempPath string, setup *setup) error {
	cmd := exec.Command("curl", "-OL",
		fmt.Sprintf("%s/%s/%s",
			os.Getenv(kubeconfigRepoUrlKey),
			setup.Deployment, getKubeConfigFilename()))
	cmd.Dir = tempPath
	l.InfoWith("Fetching kubconfig", "cmd", cmd)
	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}

func filter(commits []Commit, fn func(Commit) []string) []setup {
	index := make(map[string]bool)
	for _, commit := range commits {
		for _, entry := range fn(commit) {
			l.InfoWith("Checking commits", "entry", entry)
			if !strings.Contains(entry, getIgnorePart()) {
				index[strings.Join(strings.Split(entry, "/")[0:getSetupNameSize()+2], "/")] = true
			}
		}
	}

	l.InfoWith("collected commits", "commits", index)
	result := make([]setup, len(index))
	i := 0
	for key := range index {
		result[i] = buildSetup(key)
		i++
	}
	return result
}

func filterAdded(commits []Commit) []setup {
	return filter(commits, func(commit Commit) []string {
		return commit.Added
	})
}

func filterModified(commits []Commit) []setup {
	return filter(commits, func(commit Commit) []string {
		return commit.Modified
	})
}

func filterRemoved(commits []Commit) []setup {
	return filter(commits, func(commit Commit) []string {
		return commit.Removed
	})
}
