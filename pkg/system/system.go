package system

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"golang.org/x/sync/errgroup"
)

type HealthFunc func(ctx context.Context) error
type CleanupFunc func(ctx context.Context) error

type HealthCheckable interface {
	HealthChecks() (name string, check HealthFunc)
}

type cleanupEntry struct {
	name string
	fn   CleanupFunc
}

// System encapsulates the Useful Work that a service performs. It models this
// as a list of long-lived concurrent services (e.g an API Server, a reconciler,
// etc), and coordinates their health checks, alongside cleanup and shutdown
// signaling.
type System struct {
	services     []func(context.Context) error
	healthChecks []HealthCheckable
	cleanups     []cleanupEntry
}

// New create a new, empty system.
func New() *System {
	return &System{}
}

// Run runs any services added to the system.
//
// Run is blocking and will only return when all it's services have finished.
// The error returned will be the first error returned from any of the services.
// The terminationDelay passed in is the amount of time to wait between receiving a
// signal and cancelling the system context
func (r *System) Run(ctx context.Context, terminationDelay time.Duration) error {
	g, ctx := errgroup.WithContext(ctx)
	g.Go(func() error {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
		select {
		case <-quit:
			time.Sleep(terminationDelay)
			return fmt.Errorf("terminated")
		case <-ctx.Done():
			return nil
		}
	})

	for _, f := range r.services {
		// Capture the fn, so we don't overwrite it when starting in parallel.
		f := f
		g.Go(func() error {
			return f(ctx)
		})
	}

	return g.Wait()
}

// AddService adds the service function to the list of coordinated services.
// Once the system Run is called each service function will be invoked.
//
// The context passed into each service can be used to coordinate graceful shutdown.
// Each service should monitor the context for cancellation then stop taking on new work,
// and drain before returning.
//
// If a service needs to do any extra work to gracefully terminate, it must do
// so using a `cleanup` function.
//
// If a service depends on other services or utilities (such as a database connection) to complete
// in-flight work then the depended upon systems should remain active enough during a context
// cancellation, and only full shut down via a cleanup function.
func (r *System) AddService(s func(ctx context.Context) error) {
	r.services = append(r.services, s)
}

func (r *System) AddHealthCheck(h HealthCheckable) {
	r.healthChecks = append(r.healthChecks, h)
}

// AddCleanup stores function in the system that will be called when Cleanup is called.
// The functions added here will be invoked when Cleanup is called, which is typically.
// after Run has returned.
// The Name provided needs to be unique.
func (r *System) AddCleanup(name string, c CleanupFunc) {
	r.cleanups = append(r.cleanups, cleanupEntry{name: name, fn: c})
}

// HealthChecks returns the list of previously stored health checkers. This list can
// be used to report on the liveness and readiness of the system.
func (r *System) HealthChecks() []HealthCheckable {
	return r.healthChecks
}

// Cleanup calls each function previously added with AddCleanup. It is expected to
// be called after Run has returned to do any final work to allow the system to exit gracefully.
// Cleanup funcs are called in reverse of the order they're added in.
func (r *System) Cleanup(ctx context.Context) {
	inReverseOrder(r.cleanups, func(e cleanupEntry) {
		_ = e.fn(ctx)
	})
}

func inReverseOrder[L ~[]E, E any](list L, eachFunc func(v E)) {
	for idx := len(list) - 1; idx >= 0; idx-- {
		eachFunc(list[idx])
	}
}
