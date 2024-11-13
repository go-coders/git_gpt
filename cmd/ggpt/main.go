package main

import (
	"context"
	"flag"
	"log"
	"runtime/debug"

	"github.com/go-coders/git_gpt/internal/app"
	"github.com/go-coders/git_gpt/internal/config"
	"github.com/go-coders/git_gpt/internal/version"
	"github.com/go-coders/git_gpt/pkg/utils"
)

var (
	debugMode  = flag.Bool("debug", false, "Enable debug mode")
	configPath = flag.String("config", "", "Path to config file")
)

func main() {
	flag.Parse()

	initVersion()

	logger := utils.NewLogger(*debugMode)
	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	application, err := app.New(app.Options{
		Config:  cfg,
		Logger:  logger,
		Version: version.Version,
	})
	if err != nil {
		log.Fatalf("Failed to initialize application: %v", err)
	}

	ctx := context.Background()
	if err := application.Run(ctx); err != nil {
		log.Fatalf("Application Run failed: %v", err)
	}
}

func initVersion() {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return
	}

	version.InitFromBuildInfo(info)
}
