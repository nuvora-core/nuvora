package web

import (
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"
)

func newHandler(t *testing.T) http.Handler {
	t.Helper()

	h, err := Handler()
	if err != nil {
		t.Skip("frontend not built:", err)
	}

	return h
}

func TestHandlerServesIndex(t *testing.T) {
	h := newHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("GET / status = %d, want %d", rec.Code, http.StatusOK)
	}

	body := rec.Body.String()

	if !strings.Contains(body, `<div id="root">`) {
		t.Error("GET / body does not contain the app root")
	}

	if !strings.Contains(body, "Nuvora") {
		t.Error("GET / body does not contain the app title")
	}
}

func TestHandlerSPAFallback(t *testing.T) {
	h := newHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/some/client/route", nil)
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("GET /some/client/route status = %d, want %d", rec.Code, http.StatusOK)
	}

	if ct := rec.Header().Get("Cache-Control"); ct != "no-cache" {
		t.Errorf("GET /some/client/route Cache-Control = %q, want %q", ct, "no-cache")
	}
}

func TestHandlerServesBuiltAssets(t *testing.T) {
	h := newHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	assets := regexp.MustCompile(`/assets/[^"]+`).FindAllString(rec.Body.String(), -1)

	if len(assets) == 0 {
		t.Fatal("index.html does not reference any built assets")
	}

	for _, asset := range assets {
		req := httptest.NewRequest(http.MethodGet, asset, nil)
		rec := httptest.NewRecorder()

		h.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("GET %s status = %d, want %d", asset, rec.Code, http.StatusOK)
		}
	}
}

func TestHandlerServesAsset(t *testing.T) {
	h := newHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/favicon.svg", nil)
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("GET /favicon.svg status = %d, want %d", rec.Code, http.StatusOK)
	}

	if rec.Body.Len() == 0 {
		t.Error("GET /favicon.svg body is empty")
	}
}
