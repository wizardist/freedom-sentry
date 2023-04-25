package main

import (
	"freedom-sentry/app"
	"freedom-sentry/config"
	"log"
	"net/http"
)
import _ "net/http/pprof"

func main() {
	config.InitFlags()

	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	a := app.NewApp(app.WithDryMode(false))
	a.Run()
}
