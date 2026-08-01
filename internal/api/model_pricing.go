package api

import (
	"context"
	"net/http"

	"sub2api-usage-keeper/internal/modelsdev"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type OfficialPricingProvider interface {
	Get(context.Context) (modelsdev.Response, error)
}

func registerModelPricingRoutes(group *gin.RouterGroup, provider OfficialPricingProvider) {
	if provider == nil {
		return
	}
	group.GET("/sub2api/model-pricing", func(c *gin.Context) {
		response, err := provider.Get(c.Request.Context())
		if err != nil {
			logrus.WithError(err).Warn("load official model pricing")
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"error": "official model pricing is temporarily unavailable",
			})
			return
		}
		c.JSON(http.StatusOK, response)
	})
}
