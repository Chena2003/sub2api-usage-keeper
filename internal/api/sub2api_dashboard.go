package api

import (
	"context"
	"net/http"
	"strconv"

	"sub2api-usage-keeper/internal/quota"
	"sub2api-usage-keeper/internal/sub2api"

	"github.com/gin-gonic/gin"
)

type Sub2APIDashboardProvider interface {
	Accounts(context.Context, int) ([]quota.Sub2APIAccountQuota, error)
	Overview(context.Context, int) (quota.Sub2APIOverview, error)
	Hourly(context.Context, int) ([]sub2api.UsageOverviewRow, error)
	Models(context.Context, int, int) ([]sub2api.ModelUsageRow, error)
}

func registerSub2APIDashboardRoutes(router gin.IRoutes, provider Sub2APIDashboardProvider) {
	router.GET("/sub2api/accounts", func(c *gin.Context) {
		if provider == nil {
			writeInternalError(c, "sub2api dashboard provider is not configured", nil)
			return
		}

		accounts, err := provider.Accounts(c.Request.Context(), sub2APIQueryInt(c, "days", 7))
		if err != nil {
			writeInternalError(c, "list sub2api accounts failed", err)
			return
		}

		c.JSON(http.StatusOK, gin.H{"accounts": accounts})
	})

	router.GET("/sub2api/overview", func(c *gin.Context) {
		if provider == nil {
			writeInternalError(c, "sub2api dashboard provider is not configured", nil)
			return
		}

		overview, err := provider.Overview(c.Request.Context(), sub2APIQueryInt(c, "days", 7))
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

		points, err := provider.Hourly(c.Request.Context(), sub2APIQueryInt(c, "hours", 24))
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

		models, err := provider.Models(c.Request.Context(), sub2APIQueryInt(c, "days", 7), sub2APIQueryInt(c, "limit", 20))
		if err != nil {
			writeInternalError(c, "list sub2api models failed", err)
			return
		}

		c.JSON(http.StatusOK, gin.H{"models": models})
	})
}

func sub2APIQueryInt(c *gin.Context, name string, defaultValue int) int {
	value, err := strconv.Atoi(c.Query(name))
	if err != nil || value <= 0 {
		return defaultValue
	}
	return value
}
