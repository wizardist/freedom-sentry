package main

import (
	"github.com/wizardist/freedom-sentry/app"
	"github.com/wizardist/freedom-sentry/config"
)

func main() {
	config.InitFlags()

	a := app.NewApp(app.WithDryMode(false))
	a.Run()
}
