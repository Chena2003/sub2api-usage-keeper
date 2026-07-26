package api

import (
	"bytes"
	"context"
	"io"
	"io/fs"
	"net/http"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"sub2api-usage-keeper/internal/quota"
	"sub2api-usage-keeper/internal/service"
	"sub2api-usage-keeper/internal/version"
)

const appBasePathPlaceholder = "__APP_BASE_PATH__"

type QuotaProvider interface {
	GetCachedQuota(context.Context, quota.CacheRequest) (quota.CacheResponse, error)
	Refresh(context.Context, quota.RefreshRequest) (quota.RefreshResponse, error)
	GetRefreshTask(context.Context, string) (quota.RefreshTaskResponse, error)
}

type OptionalProviders struct {
	UsageIdentity    service.UsageIdentityProvider
	Quota            QuotaProvider
	Pricing          service.PricingProvider
	Sub2APIDashboard Sub2APIDashboardProvider
}

func NewRouter(
	staticFS fs.FS,
	usageProvider service.UsageProvider,
	basePath string,
	optionalProviders ...OptionalProviders,
) *gin.Engine {
	router := gin.New()
	_ = router.SetTrustedProxies(nil)
	router.Use(gin.Recovery())
	router.Use(securityHeadersMiddleware())

	// No CORS middleware: this service is designed for same-origin deployment
	// (static frontend served from the same Gin router). If cross-origin access
	// is needed, add a CORS middleware here and restrict allowed origins.

	appGroup := router.Group(basePath)
	registerHealthRoutes(appGroup)

	apiV1 := appGroup.Group("/api/v1")
	apiV1.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	var usageIdentityProvider service.UsageIdentityProvider
	var quotaProvider QuotaProvider
	var pricingProvider service.PricingProvider
	var sub2apiDashboardProvider Sub2APIDashboardProvider
	if len(optionalProviders) > 0 {
		usageIdentityProvider = optionalProviders[0].UsageIdentity
		quotaProvider = optionalProviders[0].Quota
		pricingProvider = optionalProviders[0].Pricing
		sub2apiDashboardProvider = optionalProviders[0].Sub2APIDashboard
	}

	dashboardRoutes := apiV1.Group("")
	registerStatusRoutes(dashboardRoutes)
	registerUsageOverviewRoute(dashboardRoutes, usageProvider)
	registerUsageAnalysisRoute(dashboardRoutes, usageProvider)
	registerUsageEventsRoute(dashboardRoutes, usageProvider, usageIdentityProvider)
	registerUsageIdentityRoutes(dashboardRoutes, usageIdentityProvider)
	registerPricingRoutes(dashboardRoutes, pricingProvider)
	registerQuotaRoutes(dashboardRoutes, quotaProvider)
	registerSub2APIDashboardRoutes(dashboardRoutes, sub2apiDashboardProvider)

	if staticFS != nil {
		if indexFile, err := staticFS.Open("index.html"); err == nil {
			_ = indexFile.Close()
			httpFS := http.FS(staticFS)
			serveIndex := func(c *gin.Context) {
				indexHTML, err := renderIndexHTML(staticFS, basePath)
				if err != nil {
					c.Status(http.StatusNotFound)
					return
				}
				setHTMLCacheHeaders(c)
				c.Data(http.StatusOK, "text/html; charset=utf-8", indexHTML)
			}
			serveAsset := func(c *gin.Context) {
				assetPath := "assets/" + strings.TrimPrefix(c.Param("filepath"), "/")
				if assetFile, err := staticFS.Open(assetPath); err == nil {
					_ = assetFile.Close()
					setStaticAssetCacheHeaders(c)
					c.FileFromFS(assetPath, httpFS)
					return
				}
				c.Status(http.StatusNotFound)
			}

			appGroup.GET("/", serveIndex)
			appGroup.GET("/assets/*filepath", serveAsset)
			appGroup.HEAD("/assets/*filepath", serveAsset)
			router.NoRoute(func(c *gin.Context) {
				requestPath, ok := stripBasePath(basePath, c.Request.URL.Path)
				if !ok {
					c.Status(http.StatusNotFound)
					return
				}
				if strings.HasPrefix(requestPath, "/api/") {
					c.Status(http.StatusNotFound)
					return
				}

				if assetPath, ok := staticAssetPath(requestPath); ok {
					if assetFile, err := staticFS.Open(assetPath); err == nil {
						_ = assetFile.Close()
						setStaticAssetCacheHeaders(c)
						c.FileFromFS(assetPath, httpFS)
						return
					}
				}

				serveIndex(c)
			})
		}
	}

	return router
}

// securityHeadersMiddleware sets a minimal set of security response headers.
// A Content-Security-Policy is intentionally omitted: the SPA relies on inline
// styles and canvas-rendered charts, and a restrictive CSP would break it.
func securityHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("Referrer-Policy", "no-referrer")
		c.Next()
	}
}

func setHTMLCacheHeaders(c *gin.Context) {
	c.Header("Cache-Control", "no-store")
	c.Header("Pragma", "no-cache")
	c.Header("Expires", "0")
}

func setStaticAssetCacheHeaders(c *gin.Context) {
	c.Header("Cache-Control", "public, max-age=31536000, immutable")
}

func renderIndexHTML(staticFS fs.FS, basePath string) ([]byte, error) {
	indexFile, err := staticFS.Open("index.html")
	if err != nil {
		return nil, err
	}
	defer indexFile.Close()
	indexHTML, err := io.ReadAll(indexFile)
	if err != nil {
		return nil, err
	}

	return bytes.ReplaceAll(
		indexHTML,
		[]byte(strconv.Quote(appBasePathPlaceholder)),
		[]byte(strconv.Quote(basePath)),
	), nil
}

func cleanURLPath(requestPath string) string {
	cleaned := path.Clean(requestPath)
	if cleaned == "." {
		return "/"
	}
	if !strings.HasPrefix(cleaned, "/") {
		return "/" + cleaned
	}
	return cleaned
}

func staticAssetPath(requestPath string) (string, bool) {
	cleaned := cleanURLPath(requestPath)
	if strings.Contains(cleaned, "\\") {
		return "", false
	}
	relPath := strings.TrimPrefix(cleaned, "/")
	if relPath == "" {
		return "", false
	}
	return relPath, true
}

func stripBasePath(basePath, requestPath string) (string, bool) {
	cleaned := cleanURLPath(requestPath)
	if basePath == "" {
		return cleaned, true
	}
	if cleaned == basePath {
		return "/", true
	}
	if !strings.HasPrefix(cleaned, basePath+"/") {
		return "", false
	}
	trimmed := strings.TrimPrefix(cleaned, basePath)
	if trimmed == "" {
		return "/", true
	}
	return trimmed, true
}

type statusResponse struct {
	Running   bool       `json:"running"`
	Timezone  string     `json:"timezone"`
	Version   string     `json:"version"`
	LastRunAt *time.Time `json:"last_run_at,omitempty"`
}

func registerStatusRoutes(router gin.IRoutes) {
	router.GET("/status", func(c *gin.Context) {
		c.JSON(http.StatusOK, statusResponse{
			Running:  true,
			Timezone: time.Local.String(),
			Version:  version.Version,
		})
	})
}
