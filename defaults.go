package main

import (
	"os"
	"strconv"
)

const (
	kubeconfigFilenameKey     = "GITOPS_KUBECONFIG_FILENAME"
	defaultKubeconfigFilename = "kubeconfig.yaml"
	setupNameKey              = "GITOPS_SETUP_NAME_SIZE"
	defaultSetupNameSize      = 2
	shellExecutableKey        = "GITOPS_SHELL_EXEC"
	defaultShellExecutable    = "/bin/bash"
)

func getSetupNameSize() int {
	set := os.Getenv(setupNameKey)
	val, err := strconv.Atoi(set)
	if set == "" || err != nil {
		return defaultSetupNameSize
	}
	return val
}

func getKubeConfigFilename() string {
	filename := os.Getenv(kubeconfigFilenameKey)
	if filename == "" {
		return defaultKubeconfigFilename
	}
	return filename
}

func getShellExecutable() string {
	sh := os.Getenv(shellExecutableKey)
	if sh == "" {
		return defaultShellExecutable
	}
	return sh
}

func getIgnorePart() string {
	return "values"
}