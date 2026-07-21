package router

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"authmicro/pkg/logger"
)

type fakePinger struct{ err error }

func (f fakePinger) PingContext(context.Context) error { return f.err }

func TestHealthEndpoints(t *testing.T) {
	tests := []struct {
		name   string
		path   string
		ping   error
		status int
	}{
		{name: "live", path: "/live", status: http.StatusOK},
		{name: "ready", path: "/ready", status: http.StatusOK},
		{name: "dependency down", path: "/ready", ping: errors.New("database unavailable"), status: http.StatusServiceUnavailable},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			r := NewRouter(nil, nil, logger.NewLogger(), fakePinger{err: test.ping})
			request := httptest.NewRequest(http.MethodGet, test.path, nil)
			response := httptest.NewRecorder()
			r.Handler().ServeHTTP(response, request)
			if response.Code != test.status {
				t.Fatalf("got status %d, want %d", response.Code, test.status)
			}
		})
	}
}
