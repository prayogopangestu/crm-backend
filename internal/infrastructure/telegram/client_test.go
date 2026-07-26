package telegram

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClientSend(t *testing.T) {
	var called bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		if r.URL.Path != "/bottoken/sendMessage" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Fatalf("unexpected content type %s", r.Header.Get("Content-Type"))
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewWithBaseURL(server.URL, server.Client())
	if err := client.Send(context.Background(), "token", "chat", "hello"); err != nil {
		t.Fatal(err)
	}
	if !called {
		t.Fatal("expected Telegram endpoint to be called")
	}
}

func TestClientReportsFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer server.Close()
	client := NewWithBaseURL(server.URL, server.Client())
	if err := client.Send(context.Background(), "token", "chat", "hello"); err == nil {
		t.Fatal("expected non-2xx response to fail")
	}
}
