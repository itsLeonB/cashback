package http

import (
	"github.com/gin-gonic/gin"
	"github.com/itsLeonB/cashback/internal/core/config"
	"github.com/itsLeonB/cashback/internal/core/logger"
	"github.com/kroma-labs/sentinel-go/httpserver"
	sentinelGin "github.com/kroma-labs/sentinel-go/httpserver/adapters/gin"
)

func setupSentinel(router *gin.Engine, skipPaths []string) {
	corsCfg := httpserver.CORSConfig{
		AllowedOrigins: config.Global.ClientUrls,
		AllowedMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{
			"Origin",
			"Content-Type",
			"Accept",
			"Authorization",
			"X-Requested-With",
			"Content-Length",
			"Accept-Encoding",
			"X-CSRF-Token",
			"Origin",
			"Cache-Control",
			"Referer",
			"User-Agent",
			"range",
			"DNT",
			"sec-ch-ua",
			"sec-ch-ua-platform",
			"sec-ch-ua-mobile",
		},
		ExposedHeaders:   []string{"Content-Length", "X-Total-Count"},
		AllowCredentials: true,
	}

	metricsCfg := httpserver.DefaultMetricsConfig()
	metricsCfg.SkipPaths = skipPaths

	metrics, err := httpserver.NewMetrics(metricsCfg)
	if err != nil {
		logger.Error(err)
	}

	tracingCfg := httpserver.DefaultTracingConfig()
	tracingCfg.SkipPaths = skipPaths

	router.Use(
		sentinelGin.RequestID(),
		sentinelGin.Timeout(config.Global.Timeout),
		sentinelGin.CORS(corsCfg),
		sentinelGin.Logger(httpserver.LoggerConfig{
			Logger:    logger.Global,
			SkipPaths: []string{"/ping", "/livez", "/readyz", "/metrics"},
		}),
		sentinelGin.Metrics(metrics),
		sentinelGin.RateLimit(httpserver.DefaultRateLimitConfig()),
		sentinelGin.Recovery(logger.Global),
		sentinelGin.Tracing(tracingCfg),
	)
}
