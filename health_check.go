package main

import (
	"strings"

	"github.com/nuclio/nuclio-sdk"
)

const (
	awsELBHealthCheckUserAgent = "ELB-HealthChecker"
)

func IsHealthCheckRequest(event nuclio.Event) bool {
	return strings.HasPrefix(event.GetHeaderString("User-Agent"), awsELBHealthCheckUserAgent) ||
		event.GetHeaderString(githubEventHeader) == "ping"

}
