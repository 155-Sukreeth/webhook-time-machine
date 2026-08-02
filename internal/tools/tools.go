// package tools tracks developer tool dependencies for CI/CD and Task automation.
// Standard Go idiom for tracking tool dependencies via tools.go.
//
//go:build tools
// +build tools

package tools

import (
	_ "github.com/go-task/task/v3/cmd/task"
	_ "github.com/golangci/golangci-lint/cmd/golangci-lint"
	_ "github.com/securego/gosec/v2/cmd/gosec"
	_ "golang.org/x/vuln/cmd/govulncheck"
)
