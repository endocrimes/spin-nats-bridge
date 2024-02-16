package httpserver

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/endocrimes/spin-nats-bridge/pkg/system"
	"golang.org/x/sync/errgroup"
)

type HTTPServer struct {
	listener    net.Listener
	server      *http.Server
	gracePeriod time.Duration
}

type Config struct {
	// Name is the name of the server for observability tools
	Name string

	// Addr is the address to listen on
	Addr string

	// Handler is the  HTTP handler to delegate requests to.
	Handler http.Handler

	// Optional
	// Network must be "tcp", "tcp4", "tcp6", "unix", "unixpacket" or "" (which defaults to tcp).
	Network string

	// ShutdownGracePeriod is the period during which the server allows requests to be fully served.
	ShutdownGracePeriod time.Duration
}

func New(ctx context.Context, cfg Config) (s *HTTPServer, err error) {
	if cfg.Network == "" {
		cfg.Network = "tcp"
	}
	ln, err := net.Listen(cfg.Network, cfg.Addr)
	if err != nil {
		return nil, err
	}
	grace := cfg.ShutdownGracePeriod
	if grace == 0 {
		grace = 10 * time.Second
	}

	return &HTTPServer{
		listener: ln,
		server: &http.Server{
			Addr:         cfg.Addr,
			Handler:      cfg.Handler,
			ReadTimeout:  55 * time.Second,
			WriteTimeout: 55 * time.Second,
		},
		gracePeriod: grace,
	}, nil
}

func (s *HTTPServer) Serve(ctx context.Context) error {
	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		// Shutdown the HTTP Server with the grace period timeout when the context
		// is cancelled.
		<-ctx.Done()
		cctx, cancel := context.WithTimeout(context.Background(), s.gracePeriod)
		defer cancel()
		if err := s.server.Shutdown(cctx); err != nil {
			return fmt.Errorf("server shutdown failed: %w", err)
		}
		return nil
	})

	g.Go(func() error {
		err := s.server.Serve(s.listener)
		if !errors.Is(err, http.ErrServerClosed) {
			return err
		}
		return nil
	})

	return g.Wait()
}

func (s HTTPServer) Addr() string {
	return s.listener.Addr().String()
}

func LoadIntoSystem(ctx context.Context, cfg Config, sys *system.System) (*HTTPServer, error) {
	server, err := New(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("error starting %q server", cfg.Name)
	}

	sys.AddService(server.Serve)
	return server, nil
}
