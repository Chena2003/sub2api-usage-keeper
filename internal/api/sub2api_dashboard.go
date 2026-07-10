package api

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	"sub2api-usage-keeper/internal/quota"
	"sub2api-usage-keeper/internal/sub2api"

	"github.com/gin-gonic/gin"
)

type Sub2APIDashboardProvider interface {
	Accounts(context.Context, time.Time) ([]quota.Sub2APIAccountQuota, error)
	AccountQuotas(context.Context, time.Time) ([]quota.Sub2APIAccountQuota, error)
	Overview(context.Context, int) (quota.Sub2APIOverview, error)
	OverviewByHours(context.Context, int) (quota.Sub2APIOverview, error)
	Hourly(context.Context, int) ([]sub2api.UsageOverviewRow, error)
	Models(context.Context, time.Time, int) ([]sub2api.ModelUsageRow, error)
	Events(context.Context, time.Time, int, int) (quota.Sub2APIEventsResponse, error)
	Rankings(context.Context, string, time.Time, int) ([]quota.Sub2APIRankingRow, error)
	RankingTrend(context.Context, string, time.Time, int) ([]quota.Sub2APIRankingTrendPoint, string, error)
	ServiceHealth(context.Context, int) (quota.Sub2APIServiceHealth, error)
}

func registerSub2APIDashboardRoutes(router gin.IRoutes, provider Sub2APIDashboardProvider) {
	router.GET("/sub2api/accounts", func(c *gin.Context) {
		if provider == nil {
			writeInternalError(c, "sub2api dashboard provider is not configured", nil)
			return
		}

		accounts, err := provider.Accounts(c.Request.Context(), sub2APISinceTimeQuery(c))
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

		accounts, err := provider.AccountQuotas(c.Request.Context(), sub2APISinceTimeQuery(c))
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

		var overview quota.Sub2APIOverview
		var err error
		if hoursStr := c.Query("hours"); hoursStr != "" {
			hours, _ := strconv.Atoi(hoursStr)
			overview, err = provider.OverviewByHours(c.Request.Context(), hours)
		} else {
			overview, err = provider.Overview(c.Request.Context(), sub2APIDaysQuery(c))
		}
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

		models, err := provider.Models(c.Request.Context(), sub2APISinceTimeQuery(c), sub2APILimitQuery(c, 20))
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

		events, err := provider.Events(c.Request.Context(), sub2APISinceTimeQuery(c), sub2APIPageQuery(c), sub2APILimitQuery(c, 100))
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

		rankings, err := provider.Rankings(c.Request.Context(), normalizeRankingDimension(c.Query("dimension")), sub2APISinceTimeQuery(c), sub2APILimitQuery(c, 20))
		if err != nil {
			writeInternalError(c, "list sub2api rankings failed", err)
			return
		}

		c.JSON(http.StatusOK, gin.H{"rankings": rankings})
	})

	router.GET("/sub2api/rankings-trend", func(c *gin.Context) {
		if provider == nil {
			writeInternalError(c, "sub2api dashboard provider is not configured", nil)
			return
		}

		points, granularity, err := provider.RankingTrend(c.Request.Context(), normalizeRankingDimension(c.Query("dimension")), sub2APISinceTimeQuery(c), sub2APILimitQuery(c, 12))
		if err != nil {
			writeInternalError(c, "get sub2api ranking trend failed", err)
			return
		}

		c.JSON(http.StatusOK, gin.H{"points": points, "granularity": granularity})
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

// sub2APISinceQuery returns a since timestamp from the hours or days query parameter.
// It accepts hours (sub-day precision) or days (backward compatible). When neither is
// present it defaults to 7 days. The maximum lookback is 90 days.
func sub2APISinceQuery(c *gin.Context) time.Time {
	now := time.Now()
	if hoursStr := c.Query("hours"); hoursStr != "" {
		if hours, err := strconv.Atoi(hoursStr); err == nil && hours > 0 {
			hours = sub2api.ClampDashboardHours(hours)
			return now.Add(-time.Duration(hours) * time.Hour)
		}
	}
	days := sub2api.ClampDashboardDays(sub2APIQueryInt(c, "days", 7))
	return now.AddDate(0, 0, -days)
}

// sub2APIUntilQuery returns an until timestamp from the until query parameter.
// It accepts an RFC3339 timestamp. When not present or invalid it returns time.Time{}
// (zero), which downstream code interprets as "now".
func sub2APIUntilQuery(c *gin.Context) time.Time {
	if untilStr := c.Query("until"); untilStr != "" {
		if parsed, err := time.Parse(time.RFC3339, untilStr); err == nil {
			return parsed
		}
	}
	return time.Time{}
}

// sub2APISinceTimeQuery returns a since timestamp from the since query parameter.
// It accepts an RFC3339 timestamp. When not present or invalid it falls back to
// sub2APISinceQuery (hours/days).
func sub2APISinceTimeQuery(c *gin.Context) time.Time {
	if sinceStr := c.Query("since"); sinceStr != "" {
		if parsed, err := time.Parse(time.RFC3339, sinceStr); err == nil {
			return parsed
		}
	}
	return sub2APISinceQuery(c)
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
