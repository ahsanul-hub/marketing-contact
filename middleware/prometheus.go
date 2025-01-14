package middleware

import (
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	RequestCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "myapp_requests_total",
			Help: "Total number of requests processed by the MyApp web server.",
		},
		[]string{"path", "status"},
	)

	ErrorCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "myapp_requests_errors_total",
			Help: "Total number of error requests processed by the MyApp web server.",
		},
		[]string{"path", "status"},
	)
)

// PrometheusInit initializes the Prometheus metrics
func PrometheusInit() {
	prometheus.MustRegister(RequestCount)
	prometheus.MustRegister(ErrorCount)
}

// TrackMetrics is a middleware that tracks request metrics
func TrackMetrics() fiber.Handler {
	return func(c *fiber.Ctx) error {
		path := c.Path()
		// Process the request
		err := c.Next()
		status := c.Response().StatusCode()

		// Increment the request count
		RequestCount.WithLabelValues(path, http.StatusText(status)).Inc()

		// Increment the error count if the status code indicates an error
		if status >= 400 {
			ErrorCount.WithLabelValues(path, http.StatusText(status)).Inc()
		}

		return err
	}
}
