package api

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"sub2api-usage-keeper/internal/quota"
	"sub2api-usage-keeper/internal/sub2api"

	"github.com/gin-gonic/gin"
)

type Sub2APIDashboardProvider interface {
	Accounts(context.Context, int) ([]quota.Sub2APIAccountQuota, error)
	AccountQuotas(context.Context, int) ([]quota.Sub2APIAccountQuota, error)
	Overview(context.Context, int) (quota.Sub2APIOverview, error)
	Hourly(context.Context, int) ([]sub2api.UsageOverviewRow, error)
	Models(context.Context, int, int) ([]sub2api.ModelUsageRow, error)
	Events(context.Context, int, int) (quota.Sub2APIEventsResponse, error)
	Rankings(context.Context, string, int, int) ([]quota.Sub2APIRankingRow, error)
	ServiceHealth(context.Context, int) (quota.Sub2APIServiceHealth, error)
}

func registerSub2APIDashboardRoutes(router gin.IRoutes, provider Sub2APIDashboardProvider) {
	router.GET("/sub2api/accounts", func(c *gin.Context) {
		if provider == nil {
			writeInternalError(c, "sub2api dashboard provider is not configured", nil)
			return
		}

		accounts, err := provider.Accounts(c.Request.Context(), sub2APIDaysQuery(c))
		if err != nil {
			writeInternalError(c, "list sub2api accounts failed", err)
			return
		}

		c.JSON(http.StatusOK, gin.H{"accounts": accounts})
	})

	router.GET("/sub2api/account-quotas", func(c *gin.Context) {
		if provider == nil {
			writeInternalError(c, "sub2api dashboard provider is not configured", nil)
			return
		}

		accounts, err := provider.AccountQuotas(c.Request.Context(), sub2APIDaysQuery(c))
		if err != nil {
			writeInternalError(c, "list sub2api account quotas failed", err)
			return
		}

		c.JSON(http.StatusOK, gin.H{"accounts": accounts})
	})

	router.GET("/sub2api/overview", func(c *gin.Context) {
		if provider == nil {
			writeInternalError(c, "sub2api dashboard provider is not configured", nil)
			return
		}

		overview, err := provider.Overview(c.Request.Context(), sub2APIDaysQuery(c))
		if err != nil {
			writeInternalError(c, "get sub2api overview failed", err)
			return
		}

		c.JSON(http.StatusOK, overview)
	})

	router.GET("/sub2api/timeseries", func(c *gin.Context) {
		if provider == nil {
			writeInternalError(c, "sub2api dashboard provider is not configured", nil)
			return
		}

		points, err := provider.Hourly(c.Request.Context(), sub2APIHoursQuery(c))
		if err != nil {
			writeInternalError(c, "get sub2api timeseries failed", err)
			return
		}

		c.JSON(http.StatusOK, gin.H{"points": points})
	})

	router.GET("/sub2api/models", func(c *gin.Context) {
		if provider == nil {
			writeInternalError(c, "sub2api dashboard provider is not configured", nil)
			return
		}

		models, err := provider.Models(c.Request.Context(), sub2APIDaysQuery(c), sub2APILimitQuery(c, 20))
		if err != nil {
			writeInternalError(c, "list sub2api models failed", err)
			return
		}

		c.JSON(http.StatusOK, gin.H{"models": models})
	})

	router.GET("/sub2api/events", func(c *gin.Context) {
		if provider == nil {
			writeInternalError(c, "sub2api dashboard provider is not configured", nil)
			return
		}

		events, err := provider.Events(c.Request.Context(), sub2APIPageQuery(c), sub2APILimitQuery(c, 100))
		if err != nil {
			writeInternalError(c, "list sub2api events failed", err)
			return
		}

		c.JSON(http.StatusOK, events)
	})

	router.GET("/sub2api/rankings", func(c *gin.Context) {
		if provider == nil {
			writeInternalError(c, "sub2api dashboard provider is not configured", nil)
			return
		}

		rankings, err := provider.Rankings(c.Request.Context(), normalizeRankingDimension(c.Query("dimension")), sub2APIDaysQuery(c), sub2APILimitQuery(c, 20))
		if err != nil {
			writeInternalError(c, "list sub2api rankings failed", err)
			return
		}

		c.JSON(http.StatusOK, gin.H{"rankings": rankings})
	})

	router.GET("/sub2api/health", func(c *gin.Context) {
		if provider == nil {
			writeInternalError(c, "sub2api dashboard provider is not configured", nil)
			return
		}

		health, err := provider.ServiceHealth(c.Request.Context(), sub2APIHoursQuery(c))
		if err != nil {
			writeInternalError(c, "get sub2api health failed", err)
			return
		}

		c.JSON(http.StatusOK, health)
	})
}

func sub2APIDaysQuery(c *gin.Context) int {
	return sub2api.ClampDashboardDays(sub2APIQueryInt(c, "days", 7))
}

func sub2APIHoursQuery(c *gin.Context) int {
	return sub2api.ClampDashboardHours(sub2APIQueryInt(c, "hours", 24))
}

func sub2APILimitQuery(c *gin.Context, defaultValue int) int {
	return sub2api.ClampDashboardLimit(sub2APIQueryInt(c, "limit", defaultValue), defaultValue)
}

func sub2APIPageQuery(c *gin.Context) int {
	return sub2api.ClampDashboardPage(sub2APIQueryInt(c, "page", 1))
}

func sub2APIQueryInt(c *gin.Context, name string, defaultValue int) int {
	value, err := strconv.Atoi(c.Query(name))
	if err != nil || value <= 0 {
		return defaultValue
	}
	return value
}

func normalizeRankingDimension(dimension string) string {
	switch strings.ToLower(strings.TrimSpace(dimension)) {
	case "api_key", "model", "account":
		return strings.ToLower(strings.TrimSpace(dimension))
	default:
		return "user"
	}
}
