package setup

import (
	"net/http"
	"net/url"
	"time"

	"github.com/stretchr/testify/require"
)

// Client represents an API client
type Client struct {
	ts      *TestSetup
	httpClt *http.Client
	addr    url.URL
}

// Do sends an HTTP request and returns a response
func (clt *Client) Do(req *http.Request) *http.Response {
	req.URL.Scheme = clt.addr.Scheme
	req.URL.Host = clt.addr.Host
	resp, err := clt.httpClt.Do(req)
	require.NoError(clt.ts.T(), err)
	return resp
}

func (ts *TestSetup) newClient() *Client {
	// Initialize client
	time.Sleep(50 * time.Millisecond)
	srvaddr := ts.apiServer.Addr()
	return &Client{
		ts: ts,
		httpClt: &http.Client{
			Timeout: time.Second * 10,
		},
		addr: url.URL{
			Scheme: "http",
			Host:   srvaddr,
		},
	}
}

// Guest creates a new unauthenticated API client
func (ts *TestSetup) Guest() *Client { return ts.newClient() }
