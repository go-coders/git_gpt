package main

import (
	"context"
	"flag"
	"log"
	"runtime/debug"
	"strings"

	"github.com/go-coders/gitchat/internal/app"
	"github.com/go-coders/gitchat/internal/config"
	"github.com/go-coders/gitchat/internal/version"
	"github.com/go-coders/gitchat/pkg/utils"
)

func initVersion() {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return
	}

	for _, dep := range info.Deps {
		if dep.Path == "github.com/go-coders/gitchat" {
			version.Version = strings.TrimPrefix(dep.Version, "v")
			break
		}
	}

	// If no version found in deps, check main module
	if version.Version == "dev" && info.Main.Version != "(devel)" {
		version.Version = strings.TrimPrefix(info.Main.Version, "v")
	}

	// Get vcs information
	var revision, time string
	for _, setting := range info.Settings {
		switch setting.Key {
		case "vcs.revision":
			revision = setting.Value
		case "vcs.time":
			time = setting.Value
		}
	}

	if revision != "" {
		version.GitCommit = revision
	}
	if time != "" {
		version.BuildTime = time
	}
}

var debugs = flag.Bool("debug", false, "Enable debug mode")

func main() {
	initVersion()
	flag.Parse()
	logger := utils.NewLogger(*debugs)
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	application, err := app.New(cfg, logger, version.Version)
	if err != nil {
		log.Fatalf("Failed to initialize application: %v", err)
	}

	ctx := context.Background()

	if err := application.Run(ctx); err != nil {
		application.HandleErr(err)
	}
}
