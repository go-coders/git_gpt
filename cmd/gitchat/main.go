package main

import (
	"context"
	"flag"
	"log"
	"runtime/debug"

	"github.com/go-coders/gitchat/internal/app"
	"github.com/go-coders/gitchat/internal/config"
	"github.com/go-coders/gitchat/internal/version"
	"github.com/go-coders/gitchat/pkg/utils"
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
