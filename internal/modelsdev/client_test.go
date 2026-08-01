package modelsdev

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestClientFetchNormalizesOfficialProviders(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"openai": {
				"id": "openai",
				"name": "OpenAI upstream",
				"models": {
					"z-model": {"id": "z-model", "name": "Z Model", "cost": {"input": 2, "output": 8}},
					"a-model": {"id": "", "name": "", "status": "deprecated", "cost": {"input": 1, "output": 4}}
				}
			},
			"anthropic": {
				"id": "anthropic",
				"name": "Anthropic upstream",
				"models": {
					"claude-test": {
						"id": "claude-test",
						"name": "Claude Test",
						"cost": {
							"input": 3,
							"output": 15,
							"cache_read": 0.3,
							"cache_write": 3.75,
							"tiers": [{
								"tier": {"type": "context", "size": 200000},
								"input": 6,
								"output": 22.5,
								"cache_read": 0.6,
								"cache_write": 7.5
							}]
						}
					},
					"free-test": {"id": "free-test", "name": "Free Test", "cost": {"input": 0, "output": 0}},
					"no-cost": {"id": "no-cost", "name": "No Cost"},
					"": {"id": "", "name": "Empty ID", "cost": {"input": 1, "output": 2}}
				}
			},
			"openrouter": {
				"id": "openrouter",
				"name": "OpenRouter",
				"models": {
					"claude-test": {"id": "claude-test", "name": "Claude Test", "cost": {"input": 1, "output": 2}}
				}
			}
		}`))
	}))
	defer server.Close()

	catalog, err := NewClient(server.URL, server.Client(), DefaultMaxResponseBytes).Fetch(context.Background())
	if err != nil {
		t.Fatalf("Fetch returned error: %v", err)
	}
	if len(catalog.Providers) != 2 {
		t.Fatalf("expected two official providers, got %+v", catalog.Providers)
	}
	if catalog.Providers[0].ID != "anthropic" || catalog.Providers[0].Name != "Anthropic" {
		t.Fatalf("expected Anthropic first, got %+v", catalog.Providers[0])
	}
	if catalog.Providers[1].ID != "openai" || catalog.Providers[1].Name != "OpenAI" {
		t.Fatalf("expected OpenAI second, got %+v", catalog.Providers[1])
	}

	anthropic := catalog.Providers[0]
	if len(anthropic.Models) != 2 {
		t.Fatalf("expected priced Anthropic models only, got %+v", anthropic.Models)
	}
	if anthropic.Models[0].ID != "claude-test" || anthropic.Models[1].ID != "free-test" {
		t.Fatalf("expected models sorted by ID, got %+v", anthropic.Models)
	}
	claude := anthropic.Models[0]
	if claude.Cost.Input == nil || *claude.Cost.Input != 3 || claude.Cost.Output == nil || *claude.Cost.Output != 15 {
		t.Fatalf("unexpected Claude base cost: %+v", claude.Cost)
	}
	if claude.Cost.CacheRead == nil || *claude.Cost.CacheRead != 0.3 || claude.Cost.CacheWrite == nil || *claude.Cost.CacheWrite != 3.75 {
		t.Fatalf("unexpected Claude cache cost: %+v", claude.Cost)
	}
	if len(claude.Tiers) != 1 {
		t.Fatalf("expected one tier, got %+v", claude.Tiers)
	}
	tier := claude.Tiers[0]
	if tier.Type != "context" || tier.Size != 200000 || tier.Cost.Input == nil || *tier.Cost.Input != 6 || tier.Cost.Output == nil || *tier.Cost.Output != 22.5 || tier.Cost.CacheRead == nil || *tier.Cost.CacheRead != 0.6 || tier.Cost.CacheWrite == nil || *tier.Cost.CacheWrite != 7.5 {
		t.Fatalf("unexpected tier: %+v", tier)
	}

	free := anthropic.Models[1]
	if free.Cost.Input == nil || *free.Cost.Input != 0 || free.Cost.Output == nil || *free.Cost.Output != 0 {
		t.Fatalf("expected explicit zero prices to remain present, got %+v", free.Cost)
	}
	if free.Cost.CacheRead != nil || free.Cost.CacheWrite != nil {
		t.Fatalf("expected missing cache prices to remain nil, got %+v", free.Cost)
	}

	openAI := catalog.Providers[1]
	if openAI.Models[0].ID != "a-model" || openAI.Models[0].Name != "a-model" || openAI.Models[0].Status != "deprecated" {
		t.Fatalf("expected map-key ID fallback, name fallback and status, got %+v", openAI.Models[0])
	}
}

func TestClientFetchRejectsNonSuccessStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
	}))
	defer server.Close()

	_, err := NewClient(server.URL, server.Client(), DefaultMaxResponseBytes).Fetch(context.Background())
	if err == nil || !strings.Contains(err.Error(), "models.dev returned status 502") {
		t.Fatalf("expected status error, got %v", err)
	}
}

func TestClientFetchRejectsOversizedResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"padding":"` + strings.Repeat("x", 128) + `"}`))
	}))
	defer server.Close()

	_, err := NewClient(server.URL, server.Client(), 32).Fetch(context.Background())
	if err == nil || !strings.Contains(err.Error(), "models.dev response exceeds") {
		t.Fatalf("expected response size error, got %v", err)
	}
}

func TestClientFetchRejectsMalformedJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"anthropic":`))
	}))
	defer server.Close()

	_, err := NewClient(server.URL, server.Client(), DefaultMaxResponseBytes).Fetch(context.Background())
	if err == nil || !strings.Contains(err.Error(), "decode models.dev response") {
		t.Fatalf("expected decode error, got %v", err)
	}
}

func TestClientFetchRejectsCatalogWithoutOfficialPricing(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"openrouter":{"models":{"x":{"cost":{"input":1,"output":2}}}}}`))
	}))
	defer server.Close()

	_, err := NewClient(server.URL, server.Client(), DefaultMaxResponseBytes).Fetch(context.Background())
	if err == nil || !strings.Contains(err.Error(), "no official model pricing") {
		t.Fatalf("expected empty catalog error, got %v", err)
	}
}
