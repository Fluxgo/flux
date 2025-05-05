package metrics

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/valyala/fasthttp/fasthttpadaptor"
)


type Config struct {
	Enabled           bool   `yaml:"enabled" json:"enabled"`
	EndpointPath      string `yaml:"endpoint_path" json:"endpoint_path"`
	ExcludedRoutes    []string `yaml:"excluded_routes" json:"excluded_routes"`
	CollectProcessMetrics bool `yaml:"collect_process_metrics" json:"collect_process_metrics"`
}


func DefaultConfig() Config {
	return Config{
		Enabled:           true,
		EndpointPath:      "/metrics",
		ExcludedRoutes:    []string{"/metrics", "/health", "/ping"},
		CollectProcessMetrics: true,
	}
}


type Metrics struct {
	config          Config
	registry        *prometheus.Registry
	requestCount    *prometheus.CounterVec
	requestDuration *prometheus.HistogramVec
	responseSize    *prometheus.HistogramVec
	activeRequests  prometheus.Gauge
	appInfo         *prometheus.GaugeVec
}

// metrics collector
func New(config Config) *Metrics {
	if config.EndpointPath == "" {
		config.EndpointPath = DefaultConfig().EndpointPath
	}

	registry := prometheus.NewRegistry()

	
	m := &Metrics{
		config:   config,
		registry: registry,
		requestCount: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_requests_total",
				Help: "Total number of HTTP requests",
			},
			[]string{"method", "path", "status"},
		),
		requestDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "http_request_duration_seconds",
				Help:    "HTTP request latencies in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"method", "path"},
		),
		responseSize: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "http_response_size_bytes",
				Help:    "HTTP response sizes in bytes",
				Buckets: []float64{100, 1000, 10000, 100000, 1000000},
			},
			[]string{"method", "path", "status"},
		),
		activeRequests: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "http_active_requests",
				Help: "Number of active HTTP requests",
			},
		),
		appInfo: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "app_info",
				Help: "Application information",
			},
			[]string{"name", "version"},
		),
	}

	
	registry.MustRegister(m.requestCount)
	registry.MustRegister(m.requestDuration)
	registry.MustRegister(m.responseSize)
	registry.MustRegister(m.activeRequests)
	registry.MustRegister(m.appInfo)

	
	if config.CollectProcessMetrics {
		registry.MustRegister(prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}))
		registry.MustRegister(prometheus.NewGoCollector())
	}

	return m
}


func (m *Metrics) SetAppInfo(name, version string) {
	m.appInfo.WithLabelValues(name, version).Set(1)
}


func (m *Metrics) Handler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		
		path := c.Path()
		for _, excluded := range m.config.ExcludedRoutes {
			if path == excluded {
				return c.Next()
			}
		}

		
		m.activeRequests.Inc()
		defer m.activeRequests.Dec()

		
		start := time.Now()

		
		err := c.Next()

		
		status := fmt.Sprintf("%d", c.Response().StatusCode())
		method := c.Method()
		elapsed := time.Since(start).Seconds()

		
		routePath := c.Route().Path
		if routePath == "" {
			routePath = path
		}

		
		m.requestCount.WithLabelValues(method, routePath, status).Inc()
		m.requestDuration.WithLabelValues(method, routePath).Observe(elapsed)
		m.responseSize.WithLabelValues(method, routePath, status).Observe(float64(len(c.Response().Body())))

		return err
	}
}


func (m *Metrics) RegisterEndpoint(app *fiber.App) {
	app.Get(m.config.EndpointPath, func(c *fiber.Ctx) error {
		
		handler := fasthttpadaptor.NewFastHTTPHandler(promhttp.HandlerFor(m.registry, promhttp.HandlerOpts{}))
		handler(c.Context())
		return nil
	})
}
