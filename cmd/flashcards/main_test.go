package main

import (
	"bufio"
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"strings"
	"syscall"
	"testing"
	"time"
)

// waitTimeout is how long tests wait for something that should happen
// almost immediately.
const waitTimeout = 5 * time.Second

func TestParsePort(t *testing.T) {
	t.Parallel()

	tests := []struct {
		value   string
		want    string
		wantErr bool
	}{
		{value: "", want: "8080"},
		{value: "0", want: "0"},
		{value: "3000", want: "3000"},
		{value: "65535", want: "65535"},
		{value: "65536", wantErr: true},
		{value: "-1", wantErr: true},
		{value: "abc", wantErr: true},
		{value: " 80", wantErr: true},
	}

	for _, test := range tests {
		t.Run(test.value, func(t *testing.T) {
			t.Parallel()

			got, err := parsePort(test.value)
			if test.wantErr {
				if !errors.Is(err, errInvalidPort) {
					t.Errorf("parsePort(%q) error = %v, want %v", test.value, err, errInvalidPort)
				}

				return
			}

			if err != nil {
				t.Fatalf("parsePort(%q) error = %v", test.value, err)
			}

			if got != test.want {
				t.Errorf("parsePort(%q) = %q, want %q", test.value, got, test.want)
			}
		})
	}
}

// blockingHandler responds with 200 OK once release is closed, after
// signalling on started that a request has arrived.
func blockingHandler(started chan<- struct{}, release <-chan struct{}) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		started <- struct{}{}

		<-release
		w.WriteHeader(http.StatusOK)
	})
}

// listen returns a listener on a free local port.
func listen(t *testing.T) net.Listener {
	t.Helper()

	var lc net.ListenConfig

	ln, err := lc.Listen(t.Context(), "tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listening: %v", err)
	}

	return ln
}

// get sends a GET request to addr in the background and returns a channel
// that receives its status code, or 0 if the request fails.
func get(t *testing.T, addr string) <-chan int {
	t.Helper()

	status := make(chan int, 1)

	go func() {
		ctx := context.WithoutCancel(t.Context())

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://"+addr, http.NoBody)
		if err != nil {
			status <- 0

			return
		}

		res, err := http.DefaultClient.Do(req)
		if err != nil {
			status <- 0

			return
		}

		_ = res.Body.Close()
		status <- res.StatusCode
	}()

	return status
}

// waitForRefusal waits until connections to addr are refused.
func waitForRefusal(t *testing.T, addr string) {
	t.Helper()

	var dialer net.Dialer

	deadline := time.Now().Add(waitTimeout)
	for time.Now().Before(deadline) {
		conn, err := dialer.DialContext(t.Context(), "tcp", addr)
		if err != nil {
			return
		}

		_ = conn.Close()

		time.Sleep(10 * time.Millisecond)
	}

	t.Fatalf("connections to %s still accepted after %v", addr, waitTimeout)
}

func TestServeFinishesInFlightRequests(t *testing.T) {
	t.Parallel()

	listener := listen(t)
	addr := listener.Addr().String()
	started := make(chan struct{}, 1)
	release := make(chan struct{})

	ctx, cancel := context.WithCancel(t.Context())
	serveErr := make(chan error, 1)

	go func() {
		serveErr <- serve(ctx, listener, blockingHandler(started, release), waitTimeout)
	}()

	status := get(t, addr)

	<-started
	cancel()

	waitForRefusal(t, addr)
	close(release)

	if got := <-status; got != http.StatusOK {
		t.Errorf("in-flight request status = %d, want %d", got, http.StatusOK)
	}

	err := <-serveErr
	if err != nil {
		t.Errorf("serve() error = %v, want nil", err)
	}
}

func TestServeTimesOut(t *testing.T) {
	t.Parallel()

	listener := listen(t)
	started := make(chan struct{}, 1)
	release := make(chan struct{})

	defer close(release)

	ctx, cancel := context.WithCancel(t.Context())
	serveErr := make(chan error, 1)

	go func() {
		serveErr <- serve(ctx, listener, blockingHandler(started, release), 50*time.Millisecond)
	}()

	_ = get(t, listener.Addr().String())

	<-started
	cancel()

	select {
	case err := <-serveErr:
		if !errors.Is(err, context.DeadlineExceeded) {
			t.Errorf("serve() error = %v, want %v", err, context.DeadlineExceeded)
		}
	case <-time.After(waitTimeout):
		t.Fatalf("serve() still running %v after the shutdown timeout", waitTimeout)
	}
}

// TestRunStopsOnSIGTERM sends SIGTERM to the test process itself. That's safe
// once run has started listening, because run catches the signal by then.
// The test isn't parallel, so that no other test is running when the signal
// arrives.
//
//nolint:paralleltest // See above.
func TestRunStopsOnSIGTERM(t *testing.T) {
	stderr, stderrWriter := io.Pipe()
	lines := make(chan string, 100)

	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			lines <- scanner.Text()
		}
	}()

	getenv := func(key string) string {
		if key == "PORT" {
			return "0"
		}

		return ""
	}

	runErr := make(chan error, 1)

	go func() {
		runErr <- run(t.Context(), getenv, stderrWriter)

		_ = stderrWriter.Close()
	}()

	waitForLine(t, lines, `"msg":"listening"`, runErr)

	err := syscall.Kill(syscall.Getpid(), syscall.SIGTERM)
	if err != nil {
		t.Fatalf("sending SIGTERM: %v", err)
	}

	select {
	case err := <-runErr:
		if err != nil {
			t.Errorf("run() error = %v, want nil", err)
		}
	case <-time.After(waitTimeout):
		t.Fatalf("run() still running %v after SIGTERM", waitTimeout)
	}
}

// waitForLine waits until a line containing substr arrives on lines, failing
// the test if runErr receives an error first or it takes too long.
func waitForLine(t *testing.T, lines <-chan string, substr string, runErr <-chan error) {
	t.Helper()

	timeout := time.After(waitTimeout)

	for {
		select {
		case line := <-lines:
			if strings.Contains(line, substr) {
				return
			}
		case err := <-runErr:
			t.Fatalf("run() returned early: %v", err)
		case <-timeout:
			t.Fatalf("no log line containing %s after %v", substr, waitTimeout)
		}
	}
}
