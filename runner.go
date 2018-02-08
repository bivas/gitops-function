package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"

	"github.com/nuclio/nuclio-sdk"
)

type DeploymentMethod int

const (
	INSTALL DeploymentMethod = iota
	UPGRADE
	DELETE
)

type Runner struct {
	workingDir   string
	githubClient *GitHub

	commits []Commit
	logger  nuclio.Logger
}

func (r *Runner) runScripts(scripts []string) {
	for _, scriptPath := range scripts {
		script, err := r.githubClient.GetFile(scriptPath)
		if err != nil {
			r.logger.WarnWith("Unable to run script (skipping)", "script", scriptPath, "error", err)
			continue
		}
		cmd := exec.Command(getShellExecutable(), script)
		cmd.Dir = r.workingDir
		output, err := cmd.CombinedOutput()
		r.logger.InfoWith("Run script", "script", scriptPath, output)
		if err != nil {
			r.logger.WarnWith("Error running script", "script", scriptPath, "error", err)
		}
	}
}

func (r *Runner) runHelm(setup *setup, args ...string) error {
	cmd := exec.Command("helm", args...)
	cmd.Env = append(os.Environ(), fmt.Sprintf("KUBECONFIG=%s/%s", r.workingDir, defaultKubeconfigFilename))
	cmd.Dir = r.workingDir
	l.InfoWith("Running Helm", "cmd", cmd)
	output, err := cmd.CombinedOutput()
	if err != nil {
		l.ErrorWith("Error running helm", "cmd", cmd, "output", string(output))
		return err
	}
	return nil
}

func (r *Runner) runDeploymentPlan(method DeploymentMethod, setup *setup, plan *DeploymentPlan) error {
	if err := getKubeConfig(r.workingDir, setup); err != nil {
		return err
	}
	switch method {
	case INSTALL:
		r.runScripts(plan.Hooks.PreInstall)
		args := []string{
			"install",
			"--name", plan.Name,
			"--namespace", setup.Namespace,
		}
		if plan.Values != "" {
			valuesFile, err := r.githubClient.GetFile(plan.Values)
			if err != nil {
				return err
			}
			args = append(args, "--values", valuesFile)
		}
		if plan.Version != "" && plan.Version != "latest" {
			args = append(args, "--version", plan.Version)
		}
		args = append(args, plan.Chart)
		if err := r.runHelm(setup, args...); err != nil {
			return err
		}
		r.runScripts(plan.Hooks.PostInstall)
	case UPGRADE:
		r.runScripts(plan.Hooks.PreUpgrade)
		args := []string{
			"upgrade",
			"--namespace", setup.Namespace,
		}
		if plan.Values != "" {
			valuesFile, err := r.githubClient.GetFile(plan.Values)
			if err != nil {
				return err
			}
			args = append(args, "--values", valuesFile)
		}
		if plan.Version != "" && plan.Version != "latest" {
			args = append(args, "--version", plan.Version)
		}
		args = append(args, plan.Name, plan.Chart)
		if err := r.runHelm(setup, args...); err != nil {
			return err
		}
		r.runScripts(plan.Hooks.PostUpgrade)
	case DELETE:
		if err := r.runHelm(setup, "delete", "--purge", plan.Name); err != nil {
			return err
		}
	}
	return nil
}

func (r *Runner) runUpgrade() error {
	for _, item := range filterModified(r.commits) {
		r.logger.InfoWith("Running upgrade", "setup", item)
		fileContent, err := r.githubClient.GetFileRaw(item.FullPath)
		if err != nil {
			return err
		}
		plan, err := NewDeploymentPlan(item.ReleaseName, []byte(fileContent))
		if err != nil {
			return err
		}
		if err := r.runDeploymentPlan(UPGRADE, &item, plan); err != nil {
			return err
		}
	}
	return nil
}

func (r *Runner) runDelete() error {
	for _, item := range filterRemoved(r.commits) {
		r.logger.InfoWith("Running removed", "setup", item)
		fileContent, err := r.githubClient.GetFileRaw(item.FullPath)
		if err != nil {
			return err
		}
		plan, err := NewDeploymentPlan(item.ReleaseName, []byte(fileContent))
		if err != nil {
			return err
		}
		if err := r.runDeploymentPlan(DELETE, &item, plan); err != nil {
			return err
		}
	}
	return nil
}

func (r *Runner) runAdded() error {
	for _, item := range filterAdded(r.commits) {
		r.logger.InfoWith("Running added", "setup", item)
		fileContent, err := r.githubClient.GetFileRaw(item.FullPath)
		if err != nil {
			return err
		}
		plan, err := NewDeploymentPlan(item.ReleaseName, []byte(fileContent))
		if err != nil {
			return err
		}
		if err := r.runDeploymentPlan(INSTALL, &item, plan); err != nil {
			return err
		}
	}
	return nil
}

func (r *Runner) Run() error {
	methods := []func() error{
		r.runAdded,
		r.runUpgrade,
		r.runDelete,
	}
	r.logger.Info("Running setup changes")
	for _, method := range methods {
		if err := method(); err != nil {
			return err
		}
	}
	return nil
}

func NewRunner(payload *PushPayload, logger nuclio.Logger) (*Runner, error) {
	runner := &Runner{logger: logger, commits: payload.Commits}
	temp, err := ioutil.TempDir("", "deployment")
	if err != nil {
		return nil, err
	}
	runner.workingDir = temp
	runner.githubClient = NewGithubClient(payload.Repository.Owner.Name, payload.Repository.Name, temp)

	return runner, nil
}
