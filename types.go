package main

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v2"
)

type DeploymentPlan struct {
	Name    string
	Chart   string `yaml:"chart"`
	Version string `yaml:"version"`
	Values  string `yaml:"values"`
	Hooks   struct {
		PreInstall  []string `yaml:"pre-install"`
		PostInstall []string `yaml:"post-install"`
		PreUpgrade  []string `yaml:"pre-upgrade"`
		PostUpgrade []string `yaml:"post-upgrade"`
	} `yaml:"hooks"`
}

func NewDeploymentPlan(name string, content []byte) (*DeploymentPlan, error) {
	plan := DeploymentPlan{}
	if err := yaml.Unmarshal(content, &plan); err != nil {
		return nil, err
	}
	plan.Name = name
	return &plan, nil
}

type Commit struct {
	Added    []string `json:"added"`
	Removed  []string `json:"removed"`
	Modified []string `json:"modified"`
}

type PushPayload struct {
	Repository struct {
		Name  string `json:"name"`
		Owner struct {
			Name string `json:"name"`
		} `json:"owner"`
	} `json:"repository"`
	Commits []Commit `json:"commits"`
}

type setup struct {
	FullPath    string
	Deployment  string
	Namespace   string
	ReleaseName string
}

func (s *setup) String() string {
	return fmt.Sprintf("setup[FullPath=%s,Namespace=%s,ReleaseName=%s,Deployment=%s]",
		s.FullPath, s.Namespace, s.ReleaseName, s.Deployment)
}

func buildSetup(raw string) setup {
	parts := strings.Split(raw, "/")
	return setup{
		FullPath:    raw,
		Deployment:  strings.Join(parts[0:getSetupNameSize()], "/"),
		Namespace:   parts[getSetupNameSize()],
		ReleaseName: parts[getSetupNameSize()+1],
	}
}
