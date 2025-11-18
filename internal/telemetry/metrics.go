package telemetry

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	HttpRequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "Total number of HTTP requests",
	}, []string{"method", "path", "status"})

	HttpRequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "http_request_duration_seconds",
		Help:    "HTTP request duration in seconds",
		Buckets: prometheus.DefBuckets,
	}, []string{"method", "path"})

	EmployeesTotal = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "employees_total",
		Help: "Total number of employees",
	})

	EmployeesByStatus = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "employees_by_status",
		Help: "Number of employees by status",
	}, []string{"status"})
)

func InitMetrics() {
	// Метрики автоматически регистрируются при импорте
}

func UpdateEmployeeMetrics(stats map[string]interface{}) {
	if total, ok := stats["total"].(int); ok {
		EmployeesTotal.Set(float64(total))
	}

	if byStatus, ok := stats["by_status"].(map[string]int); ok {
		for status, count := range byStatus {
			EmployeesByStatus.WithLabelValues(status).Set(float64(count))
		}
	}
}