package main

import (
	"os"

	"github.com/voidmaindev/go-template/cmd/api/cmd"
)

// buildSHA is set at build time via ldflags (-X main.buildSHA=...).
// Falls back to "dev" for `go run` and unflagged builds.
var buildSHA = "dev"

func main() {
	// Surface the build SHA as an env var so config.Load can pick it up
	// for Sentry release tagging without main.go importing the config pkg.
	if os.Getenv("BUILD_SHA") == "" && buildSHA != "" {
		_ = os.Setenv("BUILD_SHA", buildSHA)
	}
	cmd.Execute()
}
