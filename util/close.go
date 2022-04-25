package util

import "io"

func Close(closer io.Closer) {
	// TODO: Log an error
	_ = closer.Close()
}
