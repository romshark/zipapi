package config

import (
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/pkg/errors"
)

// File represents a TOML encoded configuration file
type File struct {
	Mode Mode `toml:"mode"`
	Log  struct {
		Debug string `toml:"debug"`
		Error string `toml:"error"`
	} `toml:"log"`
	TransportHTTP struct {
		Host              string   `toml:"host"`
		KeepAliveDuration Duration `toml:"keep-alive-duration"`
		TLS               struct {
			Enabled          bool             `toml:"enabled"`
			MinVersion       TLSVersion       `toml:"min-version"`
			CertificateFile  string           `toml:"certificate-file"`
			KeyFile          string           `toml:"key-file"`
			CurvePreferences []TLSCurveID     `toml:"curve-preferences"`
			CipherSuites     []TLSCipherSuite `toml:"cipher-suites"`
		} `toml:"tls"`
	} `toml:"transport-http"`
}

func (fl *File) mode(conf *Config) error {
	if err := fl.Mode.Validate(); err != nil {
		return err
	}
	conf.Mode = fl.Mode
	return nil
}

func (fl *File) debugLog(conf *Config) error {
	var writer io.Writer
	if strings.HasPrefix(fl.Log.Debug, "stdout") {
		writer = os.Stdout
	} else if strings.HasPrefix(fl.Log.Debug, "file:") &&
		len(fl.Log.Debug) > 5 {
		// Debug log to file
		var err error
		writer, err = os.OpenFile(
			fl.Log.Debug[5:],
			os.O_WRONLY|os.O_APPEND|os.O_CREATE,
			0660,
		)
		if err != nil {
			return errors.Wrap(err, "debug log file")
		}
	} else {
		return fmt.Errorf("invalid: '%s'", fl.Log.Debug)
	}
	conf.DebugLog = log.New(
		writer,
		"DBG: ",
		log.Ldate|log.Ltime,
	)
	return nil
}

func (fl *File) errorLog(conf *Config) error {
	var writer io.Writer
	if strings.HasPrefix(fl.Log.Error, "stderr") {
		writer = os.Stdout
	} else if strings.HasPrefix(fl.Log.Error, "file:") &&
		len(fl.Log.Error) > 5 {
		// Error log to file
		var err error
		writer, err = os.OpenFile(
			fl.Log.Error[5:],
			os.O_WRONLY|os.O_APPEND|os.O_CREATE,
			0660,
		)
		if err != nil {
			return errors.Wrap(err, "error log file")
		}
	} else {
		return fmt.Errorf("invalid: '%s'", fl.Log.Error)
	}
	conf.ErrorLog = log.New(
		writer,
		"ERR: ",
		log.Ldate|log.Ltime,
	)
	return nil
}

func (fl *File) transportHTTP(conf *Config) error {
	// Host
	if len(fl.TransportHTTP.Host) < 1 {
		return nil
	}
	conf.TransportHTTP.Host = fl.TransportHTTP.Host

	// Keep-alive duration
	conf.TransportHTTP.KeepAliveDuration = time.Duration(
		fl.TransportHTTP.KeepAliveDuration,
	)

	// TLS
	if fl.TransportHTTP.TLS.Enabled {
		conf.TransportHTTP.TLS = &TransportHTTPTLS{
			Config:              &tls.Config{},
			CertificateFilePath: fl.TransportHTTP.TLS.CertificateFile,
			PrivateKeyFilePath:  fl.TransportHTTP.TLS.KeyFile,
		}

		// Min version
		conf.TransportHTTP.TLS.Config.MinVersion = uint16(
			fl.TransportHTTP.TLS.MinVersion,
		)

		// Curve preferences
		curveIDs := make(
			[]tls.CurveID,
			len(fl.TransportHTTP.TLS.CurvePreferences),
		)
		for i, curveID := range fl.TransportHTTP.TLS.CurvePreferences {
			curveIDs[i] = tls.CurveID(curveID)
		}
		conf.TransportHTTP.TLS.Config.CurvePreferences = curveIDs
		conf.TransportHTTP.TLS.Config.PreferServerCipherSuites = true

		// Cipher suites
		cipherSuites := make([]uint16, len(fl.TransportHTTP.TLS.CipherSuites))
		for i, cipherSuite := range fl.TransportHTTP.TLS.CipherSuites {
			cipherSuites[i] = uint16(cipherSuite)
		}
		conf.TransportHTTP.TLS.Config.CipherSuites = cipherSuites
	}

	return nil
}

// FromFile reads the configuration from a file
func FromFile(path string) (*Config, error) {
	var file File
	conf := &Config{}

	// Read TOML config file
	if _, err := toml.DecodeFile(path, &file); err != nil {
		return nil, errors.Wrap(err, "TOML decode")
	}

	for setterName, setter := range map[string]func(*Config) error{
		"mode":           file.mode,
		"log.debug":      file.debugLog,
		"log.error":      file.errorLog,
		"transport-http": file.transportHTTP,
	} {
		if err := setter(conf); err != nil {
			return nil, errors.Wrap(err, setterName)
		}
	}

	if err := conf.Init(); err != nil {
		return nil, err
	}

	return conf, nil
}
