package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"sub2api-usage-keeper/internal/modelsdev"

	"github.com/gin-gonic/gin"
)

type fakeOfficialPricingProvider struct {
	response modelsdev.Response
	err      error
	calls    int
}

func (f *fakeOfficialPricingProvider) Get(context.Context) (modelsdev.Response, error) {
	f.calls++
	return f.response, f.err
}

func TestModelPricingRouteReturnsOfficialPricing(t *testing.T) {
	input := 3.0
	fetchedAt := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	provider := &fakeOfficialPricingProvider{response: modelsdev.Response{
		Providers: []modelsdev.Provider{{
			ID:   "anthropic",
			Name: "Anthropic",
			Models: []modelsdev.Model{{
				ID:   "claude-test",
				Name: "Claude Test",
				Cost: modelsdev.Cost{Input: &input},
			}},
		}},
		FetchedAt: fetchedAt,
		Stale:     true,
	}}
	router := gin.New()
	registerModelPricingRoutes(router.Group("/api/v1"), provider)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/sub2api/model-pricing", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", resp.Code, resp.Body.String())
	}
	if provider.calls != 1 {
		t.Fatalf("expected provider to be called once, got %d", provider.calls)
	}
	var response modelsdev.Response
	if err := json.Unmarshal(resp.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(response.Providers) != 1 || response.Providers[0].ID != "anthropic" || response.FetchedAt != fetchedAt || !response.Stale {
		t.Fatalf("unexpected response: %+v", response)
	}
}

func TestModelPricingRouteReturnsUnavailable(t *testing.T) {
	provider := &fakeOfficialPricingProvider{err: errors.New("upstream unavailable")}
	router := gin.New()
	registerModelPricingRoutes(router.Group("/api/v1"), provider)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/sub2api/model-pricing", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected status 503, got %d: %s", resp.Code, resp.Body.String())
	}
	if resp.Body.String() != `{"error":"official model pricing is temporarily unavailable"}` {
		t.Fatalf("unexpected response body: %s", resp.Body.String())
	}
	if provider.calls != 1 {
		t.Fatalf("expected provider to be called once, got %d", provider.calls)
	}
}

func TestModelPricingRouteIsNotRegisteredWithoutProvider(t *testing.T) {
	router := NewRouter(nil, nil, "", OptionalProviders{})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/sub2api/model-pricing", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d: %s", resp.Code, resp.Body.String())
	}
}
