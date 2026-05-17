package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"sub2api-usage-keeper/internal/quota"
	"sub2api-usage-keeper/internal/sub2api"

	"github.com/gin-gonic/gin"
)

type fakeSub2APIDashboardProvider struct {
	accountsDays int
	overviewDays int
	hourlyHours  int
	modelsDays   int
	modelsLimit  int
}

func (f *fakeSub2APIDashboardProvider) Accounts(_ context.Context, days int) ([]quota.Sub2APIAccountQuota, error) {
	f.accountsDays = days
	return []quota.Sub2APIAccountQuota{
		{
			ID:          4,
			Provider:    "openai",
			AccountType: "oauth",
			DisplayName: "openai oauth #4",
			Status:      "active",
		},
	}, nil
}

func (f *fakeSub2APIDashboardProvider) Overview(_ context.Context, days int) (quota.Sub2APIOverview, error) {
	f.overviewDays = days
	return quota.Sub2APIOverview{
		AccountCount:       1,
		ActiveAccountCount: 1,
		TotalRequests:      10,
	}, nil
}

func (f *fakeSub2APIDashboardProvider) Hourly(_ context.Context, hours int) ([]sub2api.UsageOverviewRow, error) {
	f.hourlyHours = hours
	return []sub2api.UsageOverviewRow{
		{
			BucketStart:   time.Date(2026, 5, 17, 12, 0, 0, 0, time.UTC),
			TotalRequests: 10,
		},
	}, nil
}

func (f *fakeSub2APIDashboardProvider) Models(_ context.Context, days int, limit int) ([]sub2api.ModelUsageRow, error) {
	f.modelsDays = days
	f.modelsLimit = limit
	return []sub2api.ModelUsageRow{
		{
			Model:         "claude-sonnet-4-6",
			TotalRequests: 10,
		},
	}, nil
}

func TestSub2APIAccountsRoute(t *testing.T) {
	provider := &fakeSub2APIDashboardProvider{}
	router := gin.New()
	registerSub2APIDashboardRoutes(router.Group("/api/v1"), provider)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/sub2api/accounts?days=7", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.Code)
	}
	if provider.accountsDays != 7 {
		t.Fatalf("expected days 7, got %d", provider.accountsDays)
	}
	if body := resp.Body.String(); !strings.Contains(body, `"accounts":[`) || !strings.Contains(body, `"displayName":"openai oauth #4"`) {
		t.Fatalf("unexpected response body: %s", body)
	}
}

func TestSub2APIOverviewRoute(t *testing.T) {
	provider := &fakeSub2APIDashboardProvider{}
	router := gin.New()
	registerSub2APIDashboardRoutes(router.Group("/api/v1"), provider)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/sub2api/overview?days=7", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.Code)
	}
	if provider.overviewDays != 7 {
		t.Fatalf("expected days 7, got %d", provider.overviewDays)
	}
	if body := resp.Body.String(); !strings.Contains(body, `"totalRequests":10`) {
		t.Fatalf("unexpected response body: %s", body)
	}
}

func TestSub2APITimeseriesRoute(t *testing.T) {
	provider := &fakeSub2APIDashboardProvider{}
	router := gin.New()
	registerSub2APIDashboardRoutes(router.Group("/api/v1"), provider)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/sub2api/timeseries?hours=24", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.Code)
	}
	if provider.hourlyHours != 24 {
		t.Fatalf("expected hours 24, got %d", provider.hourlyHours)
	}
	if body := resp.Body.String(); !strings.Contains(body, `"points":[`) {
		t.Fatalf("unexpected response body: %s", body)
	}
}

func TestSub2APIModelsRoute(t *testing.T) {
	provider := &fakeSub2APIDashboardProvider{}
	router := gin.New()
	registerSub2APIDashboardRoutes(router.Group("/api/v1"), provider)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/sub2api/models?days=7&limit=20", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.Code)
	}
	if provider.modelsDays != 7 {
		t.Fatalf("expected days 7, got %d", provider.modelsDays)
	}
	if provider.modelsLimit != 20 {
		t.Fatalf("expected limit 20, got %d", provider.modelsLimit)
	}
	if body := resp.Body.String(); !strings.Contains(body, `"models":[`) {
		t.Fatalf("unexpected response body: %s", body)
	}
}

func TestSub2APIQueryDefaults(t *testing.T) {
	tests := []struct {
		name       string
		path       string
		assertions func(*testing.T, *fakeSub2APIDashboardProvider)
	}{
		{
			name: "accounts non-positive days defaults to seven",
			path: "/api/v1/sub2api/accounts?days=0",
			assertions: func(t *testing.T, provider *fakeSub2APIDashboardProvider) {
				t.Helper()
				if provider.accountsDays != 7 {
					t.Fatalf("expected days 7, got %d", provider.accountsDays)
				}
			},
		},
		{
			name: "timeseries negative hours defaults to twenty four",
			path: "/api/v1/sub2api/timeseries?hours=-1",
			assertions: func(t *testing.T, provider *fakeSub2APIDashboardProvider) {
				t.Helper()
				if provider.hourlyHours != 24 {
					t.Fatalf("expected hours 24, got %d", provider.hourlyHours)
				}
			},
		},
		{
			name: "models invalid days and non-positive limit default",
			path: "/api/v1/sub2api/models?days=x&limit=0",
			assertions: func(t *testing.T, provider *fakeSub2APIDashboardProvider) {
				t.Helper()
				if provider.modelsDays != 7 {
					t.Fatalf("expected days 7, got %d", provider.modelsDays)
				}
				if provider.modelsLimit != 20 {
					t.Fatalf("expected limit 20, got %d", provider.modelsLimit)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := &fakeSub2APIDashboardProvider{}
			router := gin.New()
			registerSub2APIDashboardRoutes(router.Group("/api/v1"), provider)

			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			resp := httptest.NewRecorder()
			router.ServeHTTP(resp, req)

			if resp.Code != http.StatusOK {
				t.Fatalf("expected status 200, got %d", resp.Code)
			}
			tt.assertions(t, provider)
		})
	}
}

func TestSub2APINilProviderReturnsInternalError(t *testing.T) {
	router := gin.New()
	registerSub2APIDashboardRoutes(router.Group("/api/v1"), nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/sub2api/overview", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", resp.Code)
	}
	if body := resp.Body.String(); !strings.Contains(body, `"error":"sub2api dashboard provider is not configured"`) {
		t.Fatalf("unexpected response body: %s", body)
	}
}
