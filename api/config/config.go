package config

import (
	"log"
	"os"

	"github.com/pkg/errors"
)

// Config defines the API server configurations
type Config struct {
	Mode          Mode
	TransportHTTP *TransportHTTP
	DebugLog      *log.Logger
	ErrorLog      *log.Logger
}

// Init sets defaults and validates the configurations
func (conf *Config) Init() error {
	// Use production mode by default
	if conf.Mode == "" {
		conf.Mode = ModeProduction
	}

	// Use default HTTP config
	if conf.TransportHTTP == nil {
		conf.TransportHTTP = &TransportHTTP{
			Host: "0.0.0.0:80",
		}
	}

	// Use default debug logger to stdout
	if conf.DebugLog == nil {
		conf.DebugLog = log.New(
			os.Stdout,
			"DBG: ",
			log.Ldate|log.Ltime,
		)
	}

	// Use default error logger to stderr
	if conf.ErrorLog == nil {
		conf.ErrorLog = log.New(
			os.Stderr,
			"ERR: ",
			log.Ldate|log.Ltime,
		)
	}

	// VALIDATE

	if conf.Mode == ModeProduction {
		// Ensure TLS is enabled in production
		if conf.TransportHTTP.TLS == nil {
			return errors.New(
				"TLS must be enabled on HTTP transport in production mode",
			)
		}
	}

	return nil
}
