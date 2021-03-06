package config

import (
	log "github.com/sirupsen/logrus"
	"os"
)

func init() {

	level := log.DebugLevel
	if env := os.Getenv(EnvLogLevel); len(env) != 0 {
		switch env {
		case "warn":
			fallthrough
		case "warning":
			level = log.WarnLevel
		case "err":
			fallthrough
		case "error":
			level = log.ErrorLevel
		case "info":
			level = log.InfoLevel
		default:
			level = log.DebugLevel
		}
	}

	log.SetLevel(level)
}
