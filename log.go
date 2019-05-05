package main

import (
	"os"

	"github.com/op/go-logging"
)

// GetLogger will create a logger for you
func GetLogger() *logging.Logger {
	var log = logging.MustGetLogger("main")
	var format = logging.MustStringFormatter(
		`%{color}%{time:15:04:05.000} %{shortfunc} ▶ %{level:.4s} %{id:03x}%{color:reset} %{message}`,
	)

	backend1 := logging.NewLogBackend(os.Stderr, "", 0)
	backend1Formatter := logging.NewBackendFormatter(backend1, format)
	logging.SetBackend(backend1Formatter)
	return log
}
