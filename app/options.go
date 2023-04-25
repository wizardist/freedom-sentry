package app

import "freedom-sentry/util"

func WithDryMode(isDryMode bool) util.Option[App] {
	return func(a *App) {
		a.isDryMode = isDryMode
	}
}
