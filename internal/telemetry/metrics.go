package telemetry

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Prometheus metrics
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

var metricsFile *os.File

// InitMetrics initializes metrics
func InitMetrics() {
	// Metrics are auto-registered by promauto
}

// UpdateEmployeeMetrics updates employee-related metrics
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

// SetupMetricsWriter sets up the metrics file writer
func SetupMetricsWriter(metricsDir, metricsFileName string) (*os.File, error) {
	if err := os.MkdirAll(metricsDir, 0755); err != nil {
		return nil, fmt.Errorf("не удалось создать директорию метрик: %w", err)
	}

	metricsFilePath := metricsDir + "/" + metricsFileName
	file, err := os.OpenFile(metricsFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return nil, fmt.Errorf("не удалось открыть файл метрик: %w", err)
	}

	metricsFile = file
	return file, nil
}

// WriteMetricsToFile writes current metrics to file
func WriteMetricsToFile() {
	if metricsFile == nil {
		return
	}

	metrics, err := prometheus.DefaultGatherer.Gather()
	if err != nil {
		slog.Error("Ошибка сбора метрик", "error", err)
		return
	}

	timestamp := time.Now().Format(time.RFC3339)
	metricsFile.WriteString(fmt.Sprintf("=== METRICS DUMP %s ===\n", timestamp))

	for _, metric := range metrics {
		metricsFile.WriteString(fmt.Sprintf("Metric: %s\n", metric.GetName()))
		metricsFile.WriteString(fmt.Sprintf("Help: %s\n", metric.GetHelp()))
		metricsFile.WriteString(fmt.Sprintf("Type: %v\n", metric.GetType()))

		for _, m := range metric.GetMetric() {
			var value float64

			if counter := m.GetCounter(); counter != nil {
				value = counter.GetValue()
			} else if gauge := m.GetGauge(); gauge != nil {
				value = gauge.GetValue()
			} else if histogram := m.GetHistogram(); histogram != nil {
				value = histogram.GetSampleSum()
			} else if summary := m.GetSummary(); summary != nil {
				value = summary.GetSampleSum()
			} else if untyped := m.GetUntyped(); untyped != nil {
				value = untyped.GetValue()
			}

			labels := make([]string, 0, len(m.GetLabel()))
			for _, label := range m.GetLabel() {
				labels = append(labels, fmt.Sprintf("%s=%s", label.GetName(), label.GetValue()))
			}

			labelStr := ""
			if len(labels) > 0 {
				labelStr = fmt.Sprintf(" {%s}", strings.Join(labels, ", "))
			}

			metricsFile.WriteString(fmt.Sprintf("  %s%s: %f\n", metric.GetName(), labelStr, value))
		}
		metricsFile.WriteString("\n")
	}
	metricsFile.WriteString("=== END METRICS DUMP ===\n\n")

	metricsFile.Sync()
}

// StartMetricsWriter starts periodic metrics writing
func StartMetricsWriter() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		WriteMetricsToFile()
	}
}

