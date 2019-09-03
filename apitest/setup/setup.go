package setup

import (
	"context"
	"testing"

	"zipapi/api"
	"zipapi/api/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSetup represents the Dgraph-based server setup of an individual test
type TestSetup struct {
	t         *testing.T
	apiServer api.Server
	shutdown  chan struct{}
}

// T returns the test reference
func (ts *TestSetup) T() *testing.T { return ts.t }

// APIServer returns the API server interface
func (ts *TestSetup) APIServer() api.Server { return ts.apiServer }

// New creates a new test setup
func New(t *testing.T, conf *config.Config) *TestSetup {
	if conf == nil {
		conf = &config.Config{}
	}

	// Partially override the config
	conf.Mode = config.ModeDebug
	conf.TransportHTTP = &config.TransportHTTP{
		Host: "localhost:",
	}

	apiServer, err := api.NewServer(conf)

	ts := &TestSetup{
		t:         t,
		apiServer: apiServer,
		shutdown:  make(chan struct{}),
	}

	require.NoError(t, err)
	go func() {
		assert.NoError(t, apiServer.Run())

		// Notify server shutdown
		ts.shutdown <- struct{}{}
	}()

	return ts
}

// Teardown gracefully terminates the test,
// this method MUST BE DEFERRED until the end of the test!
func (ts *TestSetup) Teardown() {
	// Stop the API server instance
	if err := ts.apiServer.Shutdown(context.Background()); err != nil {
		// Don't break on shutdown failure, remove database before quitting!
		ts.t.Errorf("API server shutdown: %s", err)
	}

	// Wait for the server to shut down
	<-ts.shutdown
}
