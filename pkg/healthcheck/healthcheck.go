package healthcheck

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/endocrimes/spin-nats-bridge/pkg/httpserver"
	"github.com/endocrimes/spin-nats-bridge/pkg/system"
	"github.com/hellofresh/health-go/v5"
)

type API struct {
	router *http.ServeMux
}

func New(ctx context.Context, checked []system.HealthCheckable) (*API, error) {
	router := http.NewServeMux()

	health, err := newHealthHandlers(checked)
	if err != nil {
		return nil, fmt.Errorf("failed to create health checks: %w", err)
	}

	router.HandleFunc("GET /zhealth", handleHealth(health))

	return &API{router: router}, nil
}

func handleHealth(h *health.Health) http.HandlerFunc {
	return func(writer http.ResponseWriter, req *http.Request) {
		ctx := context.Background()
		check := h.Measure(ctx)

		code := http.StatusOK

		switch check.Status {
		case health.StatusUnavailable, health.StatusTimeout:
			code = http.StatusServiceUnavailable
		case health.StatusPartiallyAvailable:
			// Status OK, todo: log
		}

		writer.Header().Add("Content-Type", "application/json")
		writer.WriteHeader(code)
		// We don't actually care if this fails
		bytes, _ := json.Marshal(check)
		_, _ = writer.Write(bytes)
	}
}

func newHealthHandlers(checked []system.HealthCheckable) (*health.Health, error) {
	checker, err := health.New()
	if err != nil {
		return nil, err
	}

	for _, c := range checked {
		name, check := c.HealthChecks()

		err = checker.Register(health.Config{
			Name:      name,
			Timeout:   time.Second * 5,
			SkipOnErr: false,
			Check:     health.CheckFunc(check),
		})
		if err != nil {
			return nil, err
		}
	}

	return checker, nil
}

func (a *API) Handler() http.Handler {
	return a.router
}

func LoadIntoSystem(ctx context.Context, addr string, sys *system.System) (*httpserver.HTTPServer, error) {
	healthAPI, err := New(ctx, sys.HealthChecks())
	if err != nil {
		return nil, fmt.Errorf("error creating health check API")
	}

	return httpserver.LoadIntoSystem(ctx, httpserver.Config{
		Name:    "healthchecks",
		Addr:    addr,
		Handler: healthAPI.Handler(),
	}, sys)
}
