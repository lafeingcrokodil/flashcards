// Command flashcards runs the web server for the flashcards app.
package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/lafeingcrokodil/flashcards/v3/internal/server"
	"github.com/lafeingcrokodil/flashcards/v3/web"
)

const (
	// defaultPort is the port to listen on if PORT isn't set.
	defaultPort = "8080"
	// maxPort is the highest valid TCP port.
	maxPort = 65535
	// readHeaderTimeout limits how long a client may take to send the request
	// headers, so that slow clients can't hold connections open indefinitely.
	readHeaderTimeout = 10 * time.Second
	// shutdownTimeout limits how long in-flight requests may take to finish
	// after a shutdown signal.
	shutdownTimeout = 5 * time.Second
)

var errInvalidPort = errors.New("invalid port")

func main() {
	if err := run(context.Background(), os.Getenv, os.Stderr); err != nil {
		slog.New(slog.NewJSONHandler(os.Stderr, nil)).Error("exiting", "error", err)
		os.Exit(1)
	}
}

// run starts the server and blocks until it has shut down, which it does when
// ctx is cancelled or the process receives SIGINT or SIGTERM.
func run(ctx context.Context, getenv func(string) string, stderr io.Writer) error {
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	logger := slog.New(slog.NewJSONHandler(stderr, nil))

	port, err := parsePort(getenv("PORT"))
	if err != nil {
		return err
	}

	frontend, err := web.Build()
	if err != nil {
		return fmt.Errorf("loading the front end: %w", err)
	}

	var lc net.ListenConfig

	listener, err := lc.Listen(ctx, "tcp", ":"+port)
	if err != nil {
		return fmt.Errorf("listening: %w", err)
	}

	logger.InfoContext(ctx, "listening", "address", listener.Addr().String())

	if err := serve(ctx, listener, server.New(frontend), shutdownTimeout); err != nil {
		return err
	}

	logger.InfoContext(ctx, "shut down")

	return nil
}

// parsePort returns the port to listen on, given the value of PORT.
func parsePort(value string) (string, error) {
	if value == "" {
		return defaultPort, nil
	}

	port, err := strconv.Atoi(value)
	if err != nil || port < 0 || port > maxPort {
		return "", fmt.Errorf("%w: %q", errInvalidPort, value)
	}

	return value, nil
}

// serve serves HTTP requests on listener until ctx is cancelled, then waits up to
// timeout for in-flight requests to finish. It returns an error if serving
// fails or the requests don't finish in time.
func serve(ctx context.Context, listener net.Listener, handler http.Handler, timeout time.Duration) error {
	srv := &http.Server{
		Handler:           handler,
		ReadHeaderTimeout: readHeaderTimeout,
	}

	serveErr := make(chan error, 1)

	go func() {
		serveErr <- srv.Serve(listener)
	}()

	select {
	case err := <-serveErr:
		return fmt.Errorf("serving: %w", err)
	case <-ctx.Done():
	}

	shutdownCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), timeout)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		// Close the connections that are still open, so that their requests
		// don't keep running after serve returns.
		closeErr := srv.Close()

		return fmt.Errorf("shutting down: %w", errors.Join(err, closeErr))
	}

	return nil
}
