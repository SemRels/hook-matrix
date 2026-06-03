// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 The semrel Authors

package plugin

import (
	"bytes"
	"context"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (fn roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return fn(req)
}

func TestMatrixNotifierNotifySuccess(t *testing.T) {
	var receivedMethod string
	var receivedPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedMethod = r.Method
		receivedPath = r.URL.Path
		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Error("expected Authorization Bearer header")
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"event_id":"$eventid"}`))
	}))
	defer srv.Close()

	n := NewMatrixNotifier(MatrixConfig{
		HomeserverURL: srv.URL,
		RoomID:        "!testroom:matrix.org",
		AccessToken:   "test-token",
	})

	if err := n.Notify(context.Background(), "v1.2.3", "## Changes\n- new feature", "myapp"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if receivedMethod != http.MethodPut {
		t.Errorf("expected PUT, got %s", receivedMethod)
	}
	if !strings.Contains(receivedPath, "m.room.message") {
		t.Errorf("expected m.room.message in path, got %s", receivedPath)
	}
}

func TestMatrixNotifierNotifyRetriesOnServerError(t *testing.T) {
	var attempts int
	var logs bytes.Buffer
	oldWriter := retryLogWriter
	retryLogWriter = &logs
	defer func() { retryLogWriter = oldWriter }()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 3 {
			w.WriteHeader(http.StatusBadGateway)
			_, _ = w.Write([]byte(`{"errcode":"M_UNKNOWN"}`))
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"event_id":"$eventid"}`))
	}))
	defer srv.Close()

	n := NewMatrixNotifier(MatrixConfig{
		HomeserverURL: srv.URL,
		RoomID:        "!room:matrix.org",
		AccessToken:   "token",
		MaxRetries:    2,
		RetryDelay:    time.Millisecond,
	})
	if err := n.Notify(context.Background(), "v1.0.0", "", "repo"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if attempts != 3 {
		t.Fatalf("expected 3 attempts, got %d", attempts)
	}
	if got := logs.String(); !strings.Contains(got, "retry attempt 1/2") || !strings.Contains(got, "retry attempt 2/2") {
		t.Fatalf("expected retry logs, got %q", got)
	}
}

func TestMatrixNotifierNotifyDoesNotRetryOnClientError(t *testing.T) {
	var attempts int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"errcode":"M_FORBIDDEN"}`))
	}))
	defer srv.Close()

	n := NewMatrixNotifier(MatrixConfig{
		HomeserverURL: srv.URL,
		RoomID:        "!room:matrix.org",
		AccessToken:   "bad-token",
		MaxRetries:    3,
		RetryDelay:    time.Millisecond,
	})

	err := n.Notify(context.Background(), "v1.0.0", "", "repo")
	if err == nil || !strings.Contains(err.Error(), "unexpected status 403") {
		t.Fatalf("expected 403 response error, got %v", err)
	}
	if attempts != 1 {
		t.Fatalf("expected 1 attempt, got %d", attempts)
	}
}

func TestMatrixNotifierNotifyRetriesOnNetworkError(t *testing.T) {
	var attempts int
	n := NewMatrixNotifier(MatrixConfig{
		HomeserverURL: "https://matrix.example.test",
		RoomID:        "!room:matrix.org",
		AccessToken:   "token",
		MaxRetries:    2,
		RetryDelay:    time.Millisecond,
	})
	n.client = &http.Client{
		Timeout: DefaultTimeout,
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			attempts++
			if attempts < 3 {
				return nil, &net.DNSError{Err: "temporary failure", IsTemporary: true}
			}
			return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(`{"event_id":"$x"}`)), Header: make(http.Header)}, nil
		}),
	}

	if err := n.Notify(context.Background(), "v1.0.0", "", "repo"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if attempts != 3 {
		t.Fatalf("expected 3 attempts, got %d", attempts)
	}
}

func TestMatrixNotifierDefaultsAndBuilders(t *testing.T) {
	n := NewMatrixNotifier(MatrixConfig{HomeserverURL: "https://matrix.org", RoomID: "!room:matrix.org", AccessToken: "tok"})
	if n.client.Timeout != DefaultTimeout {
		t.Fatalf("expected default timeout %s, got %s", DefaultTimeout, n.client.Timeout)
	}
	if n.cfg.MaxRetries != DefaultMaxRetries || n.cfg.RetryDelay != DefaultRetryDelay {
		t.Fatalf("unexpected retry defaults: %+v", n.cfg)
	}
	if got := buildMatrixBody("v1.0.0", strings.Repeat("y", 3000), "repo"); !strings.Contains(got, "(truncated)") {
		t.Fatalf("expected truncated body, got %q", got)
	}
	if got := buildMatrixHTML("v2.0.0", "", ""); !strings.Contains(got, "v2.0.0") {
		t.Fatalf("expected version in HTML, got %q", got)
	}
}
