# GitOps

The process is aimed to provide a simple function for updating Kubernetes cluster state using [helm](https://github.com/kubernetes/helm) as deployment manager.

**Basic Flow**

- File added to setup directory and pushed to GitHub
  1. GitHub sends a webhook of added file
  2. GitOps Function analise the webhook payload and decides to add a deployment to setup
  3. Run `pre-install` scripts
  4. Run `helm install` with the provided values
  5. Run `post-install` scripts
- File modified in setup directory and pushed to GitHub
  1. GitHub sends a webhook of modified file
  2. GitOps Function analise the webhook payload and decides to upgrade a deployment in setup
  3. Run `pre-upgrade` scripts
  4. Run `helm upgrade` with the provided values
  5. Run `post-upgrade` scripts
- File deleted from setup directory and pushed to GitHub
  1. GitHub sends a webhook of removed file
  2. GitOps Function analise the webhook payload and decides to removed a deployment from setup
  3. Run `helm delete` of provided values

**Installation**

```bash
$ nuctl deploy --runtime golang --env GITHUB_ACCESS_TOKEN=<access_token>,KUBECONFIG_REPO_URL="http://username:password@example.com" --path /path/to/gitops gitops
```

## Requirements

- Github repository holding the operational data (described below)
- Kubeconfig Store: Webserver (preferably secured one) holding the kube config yaml file with ability to access `helm` commands
- Github access token with `repo` privileges (accessing the git repo holding the cluster ops state)
- Github WebHook configured to receive `push` events

### Github Ops Repository

This repository should be your single source of truth for cluster state. Directory structure **must** be as follows:
```
- setup_1
  - namespace_1
    - plan_1
- setup_2
  - namespace_2
    - plan_2
    - plan_3
  - namespace_3
    - plan_4
```

Deployment plan file content is described below.

For example, let's assume we wish to install a **demo** deployment plan on our **aws** cluster named **foo** in the **default** namespace. Our directory structure will be as follows:

```
- aws
  - foo
    - default
      - demo
```

### Kubeconfig Store

`kubeconfig.yaml` file should be placed in the following structure: `/setup_path/kubeconfig.yaml`

For example:

Let's assume we have a setup in **aws** named **foo**, our `kubeconfig.yaml` file should be accessed via:

```
http://user:password@example.com/aws/foo/kubeconfig.yaml
```

## Deployment Plan

The file describe the `helm` chart configuration and (optional) scripts to invoke as install/upgrade hook.

Structure:
```
chart: stable/nuclio
version: 1.2
values: path/to/override/values.yaml
hooks:
    pre-install:
        - path/to/script1
    post-install:
        - path/to/script2
        - path/to/script3
    pre-upgrade:
        - path/to/script4
    post-upgrade:
        - path/to/script5
```

- `chart` (**required**): which chart to install
- `version` (optional): chart version to install (default: *latest*)
- `values` (optional): path (in repository) for chart override values
- `hooks` (optional): set of scripts to run on install or upgrade event
    - `pre-install`: run *before* installing a chart
    - `post-install`: run *after* installing a chart
    - `pre-upgrade`: run *before* upgrading a chart
    - `post-upgrade`: run *after* upgrading a chart


## Configuration

The following environment values are available to configure our function:

- `GITHUB_ACCESS_TOKEN` (**required**): Github access token with required permissions
- `KUBECONFIG_REPO_URL` (**required**): HTTP endpoint for getting the cluster's kubeconfig yaml file
- `GITOPS_KUBECONFIG_FILENAME` (optional): Name of the kubeconfig yaml file (default: `kubeconfig.yaml`
- `GITOPS_SETUP_NAME_SIZE` (optional): How many prefixed path parts are used to identify the setup name (default: `2`)
- `GITOPS_SHELL_EXEC` (optional): Shell to use when invoking the hook scripts (default: `/bin/bash`)