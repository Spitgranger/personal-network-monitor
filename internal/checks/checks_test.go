package checks

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestCheckHttpSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(server.Close)
	client := server.Client()
	ctx, close := context.WithTimeout(context.Background(), time.Second*10)
	defer close()
	result := CheckHttp(ctx, client, server.URL)
	if result.Status != "success" {
		t.Fatalf("Expected success got  %+v", result)
	}
	if result.StatusCode != http.StatusOK {
		t.Errorf("Expected Status OK got %d", result.StatusCode)
	}
	if result.ErrorCode != "" {
		t.Fatalf("Expected no error got %s", result.ErrorCode)
	}
}

func TestCheckHttpTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-time.After(time.Second * 5):
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("completed"))
		case <-r.Context().Done():
			// The client or server timed out before the delay finished
			return
		}
	}))
	t.Cleanup(server.Close)
	client := server.Client()
	ctx, close := context.WithTimeout(context.Background(), time.Second*2)
	defer close()
	result := CheckHttp(ctx, client, server.URL)
	if result.Status != "failure" {
		t.Fatalf("Expected failure got  %+v", result)
	}
	if result.ErrorCode != "Request timed out" && result.DurationMS > 0{
		t.Fatalf("Expected Request timed out error got %s", result.ErrorCode)
	}
}

func TestCheckHttpTimeoutAfterHeaderReceived(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.(http.Flusher).Flush()
		select {
		case <-time.After(time.Second * 5):
			w.Write([]byte("Completed"))
		case <-r.Context().Done():
			// The client or server timed out before the delay finished
			return
		}
	}))
	t.Cleanup(server.Close)
	client := server.Client()
	ctx, close := context.WithTimeout(context.Background(), time.Second*2)
	defer close()
	result := CheckHttp(ctx, client, server.URL)
	if result.Status != "failure" {
		t.Fatalf("Expected failure got  %+v", result)
	}
	if result.ErrorCode != "Request timed out" && result.DurationMS > 0 {
		t.Fatalf("Expected Request timed out error got %s", result.ErrorCode)
	}
}
