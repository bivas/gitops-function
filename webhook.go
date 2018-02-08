package main

import (
	"encoding/json"
	"errors"

	"github.com/nuclio/nuclio-sdk"
)

const (
	githubEventHeader = "X-Github-Event"
)

var l nuclio.Logger

func Webhook(context *nuclio.Context, event nuclio.Event) (interface{}, error) {
	if IsHealthCheckRequest(event) {
		return "ok", nil
	}
	l = context.Logger
	context.Logger.InfoWith("Checking for github header", githubEventHeader, event.GetHeaderString(githubEventHeader))
	if event.GetHeaderString(githubEventHeader) != "push" {
		return "", errors.New("not a GitHub 'push' event")
	}
	payload := &PushPayload{}
	if err := json.Unmarshal(event.GetBody(), payload); err != nil {
		return "", err
	}
	runner, err := NewRunner(payload, context.Logger)
	if err != nil {
		return "", err
	}
	if err := runner.Run(); err != nil {
		return "", err
	}
	return "Gitops Completed", nil
}
