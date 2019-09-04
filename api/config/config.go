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
	App           App
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

	// Set default file size limit to 1mb
	if conf.App.MaxFileSize == 0 {
		conf.App.MaxFileSize = 1024 * 1024
	}

	// Set default request size limit to 4mb
	if conf.App.MaxReqSize == 0 {
		conf.App.MaxReqSize = 1024 * 1024 * 4
	}

	// Set default multipart memoery buffer to 1mb
	if conf.App.MaxMultipartMembuf == 0 {
		conf.App.MaxMultipartMembuf = 1024 * 1024
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
