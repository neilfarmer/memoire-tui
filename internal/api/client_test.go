package api

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func newTestClient(handler http.Handler) (*Client, *httptest.Server) {
	srv := httptest.NewServer(handler)
	c := New(srv.URL, "pat_test")
	return c, srv
}

func TestGetJSON(t *testing.T) {
	c, srv := newTestClient(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer pat_test" {
			t.Errorf("auth header=%q", r.Header.Get("Authorization"))
		}
		_, _ = w.Write([]byte(`{"hello":"world"}`))
	}))
	defer srv.Close()
	var out struct{ Hello string }
	if err := c.Get("/x", &out); err != nil {
		t.Fatal(err)
	}
	if out.Hello != "world" {
		t.Errorf("got %q", out.Hello)
	}
}

func TestErrorBody(t *testing.T) {
	c, srv := newTestClient(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"bad input"}`))
	}))
	defer srv.Close()
	err := c.Get("/x", nil)
	var ae *APIError
	if !errors.As(err, &ae) {
		t.Fatalf("expected APIError, got %T", err)
	}
	if ae.Status != 400 || ae.Message != "bad input" {
		t.Errorf("ae=%+v", ae)
	}
}

func TestPATForbiddenSentinel(t *testing.T) {
	c, srv := newTestClient(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer srv.Close()
	err := c.Get("/tokens", nil)
	if !errors.Is(err, ErrPATForbidden) {
		t.Errorf("expected ErrPATForbidden, got %v", err)
	}
}

func TestPostBodySent(t *testing.T) {
	c, srv := newTestClient(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("content-type=%q", r.Header.Get("Content-Type"))
		}
		buf := make([]byte, r.ContentLength)
		_, _ = r.Body.Read(buf)
		if !strings.Contains(string(buf), "value") {
			t.Errorf("body=%q", string(buf))
		}
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()
	var out struct {
		OK bool `json:"ok"`
	}
	if err := c.Post("/x", map[string]string{"key": "value"}, &out); err != nil {
		t.Fatal(err)
	}
	if !out.OK {
		t.Error("expected ok=true")
	}
}

func TestQueryString(t *testing.T) {
	c, srv := newTestClient(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("q") != "hello" {
			t.Errorf("q=%q", r.URL.Query().Get("q"))
		}
	}))
	defer srv.Close()
	q := map[string][]string{"q": {"hello"}}
	if err := c.GetQ("/x", q, nil); err != nil {
		t.Fatal(err)
	}
}

func TestDelete204(t *testing.T) {
	c, srv := newTestClient(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()
	if err := c.Delete("/x/1"); err != nil {
		t.Errorf("delete: %v", err)
	}
}

// TestDoContextCancellation: a request cancelled via context returns the
// cancellation error rather than blocking until the client-level timeout.
func TestDoContextCancellation(t *testing.T) {
	c, srv := newTestClient(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Hold the connection open longer than the test will wait.
		time.Sleep(2 * time.Second)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer srv.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Millisecond)
	defer cancel()
	err := c.DoContext(ctx, http.MethodGet, "/slow", nil, nil, nil)
	if err == nil {
		t.Fatal("expected an error from cancelled context, got nil")
	}
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("expected context.DeadlineExceeded; got %v", err)
	}
}

// TestWithContextCancellation proves c.WithContext(ctx) propagates the
// caller's cancellation through the existing Do/Get/Post path without
// every call site needing to use DoContext explicitly.
func TestWithContextCancellation(t *testing.T) {
	c, srv := newTestClient(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer srv.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Millisecond)
	defer cancel()
	err := c.WithContext(ctx).Get("/slow", nil)
	if err == nil {
		t.Fatal("expected an error from cancelled context, got nil")
	}
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("expected context.DeadlineExceeded; got %v", err)
	}
}
