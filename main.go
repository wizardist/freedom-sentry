package main

import (
	"log/slog"
	"os"

	"github.com/wizardist/freedom-sentry/app"
	"github.com/wizardist/freedom-sentry/config"
)

func main() {
	config.InitFlags()

	// Initialize structured logging
	logLevel := config.GetLogLevel()
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: logLevel,
	}))
	slog.SetDefault(logger)

	slog.Info("freedom-sentry starting", "log_level", logLevel.String())

	a := app.NewApp(app.WithDryMode(false))
	a.Run()
}
