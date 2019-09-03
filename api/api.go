package api

import (
	"context"
	"net"
	"net/http"

	"github.com/romshark/zipapi/api/config"
	"github.com/romshark/zipapi/store"
	storemock "github.com/romshark/zipapi/store/mock"

	"github.com/pkg/errors"
)

// Server interfaces an API server implementation
type Server interface {
	// Run starts the API server and blocks the calling goroutine
	Run() error

	// Shutdown instructs the server to shut down gracefully and blocks until
	// the server is shut down
	Shutdown(context.Context) error

	// Addr returns the server's host address
	Addr() string

	// Store returns the server's store interface
	Store() store.Store
}

type server struct {
	conf        *config.Config
	httpSrv     *http.Server
	tcpListener net.Listener
	store       store.Store
}

// NewServer creates a new API server instance
func NewServer(conf *config.Config) (Server, error) {
	if err := conf.Init(); err != nil {
		return nil, errors.Wrap(err, "config initialization")
	}

	// Initialize API server instance
	srv := &server{
		conf: conf,
	}

	// Initialize store instance
	srv.store = storemock.New()

	if err := srv.store.Init(); err != nil {
		return nil, errors.Wrap(err, "store preparation")
	}

	// Initialize HTTP server
	srv.httpSrv = &http.Server{
		Addr:     conf.TransportHTTP.Host,
		ErrorLog: conf.ErrorLog,
		Handler:  srv,
	}
	if conf.TransportHTTP.TLS != nil {
		srv.httpSrv.TLSConfig = conf.TransportHTTP.TLS.Config.Clone()
	}

	// Initialize the TCP listener
	addr := srv.httpSrv.Addr
	if addr == "" {
		addr = ":http"
	}
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, errors.Wrap(err, "TCP listener setup")
	}
	srv.httpSrv.Addr = listener.Addr().String()

	srv.tcpListener = tcpKeepAliveListener{
		TCPListener:       listener.(*net.TCPListener),
		KeepAliveDuration: srv.conf.TransportHTTP.KeepAliveDuration,
	}

	return srv, nil
}

func (srv *server) logErrf(format string, v ...interface{}) {
	srv.conf.ErrorLog.Printf(format, v...)
}

// Launch implements the Server interface
func (srv *server) Run() error {
	// Launch the HTTP server
	if srv.conf.TransportHTTP.TLS != nil {
		srv.conf.DebugLog.Print("listening https://" + srv.httpSrv.Addr)

		if err := srv.httpSrv.ServeTLS(
			srv.tcpListener,
			srv.conf.TransportHTTP.TLS.CertificateFilePath,
			srv.conf.TransportHTTP.TLS.PrivateKeyFilePath,
		); err != http.ErrServerClosed {
			return err
		}
	} else {
		srv.conf.DebugLog.Print("listening http://" + srv.httpSrv.Addr)

		if err := srv.httpSrv.Serve(
			srv.tcpListener,
		); err != http.ErrServerClosed {
			return err
		}
	}

	return nil
}

// Shutdown implements the Server interface
func (srv *server) Shutdown(ctx context.Context) error {
	return srv.httpSrv.Shutdown(ctx)
}

// Addr implements the Server interface
func (srv *server) Addr() string { return srv.httpSrv.Addr }

// Store implements the Server interface
func (srv *server) Store() store.Store { return srv.store }

func (srv *server) ServeHTTP(out http.ResponseWriter, in *http.Request) {
	// Allow POST methods only
	if in.Method != "POST" {
		http.Error(
			out,
			http.StatusText(http.StatusMethodNotAllowed),
			http.StatusMethodNotAllowed,
		)
		return
	}

	var handler func(http.ResponseWriter, *http.Request) error

	switch in.URL.Path {
	// POST /archive
	case "/archive":
		fallthrough
	case "/archive/":
		handler = srv.postArchive
	// 404
	default:
		http.Error(
			out,
			http.StatusText(http.StatusNotFound),
			http.StatusNotFound,
		)
		return
	}

	if err := handler(out, in); err != nil {
		// Log internal errors and return '500 Internal Server Error'
		srv.conf.ErrorLog.Printf(
			"internal error: (%s '%s'): %s",
			in.URL.Path,
			in.Method,
			err,
		)
		http.Error(
			out,
			http.StatusText(http.StatusInternalServerError),
			http.StatusInternalServerError,
		)
	}
}
