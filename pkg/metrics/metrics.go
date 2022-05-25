package metrics

// package metrics 用于注册和记录 HTTP 服务一般所需要的几个指标
// 用于 prometheus 的分析

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/cast"
)

var metricsRequestsTotal *prometheus.CounterVec

var metricsRequestsCost *prometheus.HistogramVec

// InitMetrics 主动注册 Metric 指标
func InitMetrics(namespace string) {
	// metricsRequestsTotal metrics for request total 计数器（Counter）
	metricsRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "http_api_requests_total",
			Help:      "request(ms) total",
		},
		[]string{"server_name", "method", "path", "http_code", "business_code"},
	)

	// metricsRequestsCost metrics for requests cost 累积直方图（Histogram）
	metricsRequestsCost = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "http_api_requests_cost",
			Help:      "request(ms) cost milliseconds",
		},
		[]string{"server_name", "method", "path", "http_code", "business_code"},
	)

	prometheus.MustRegister(metricsRequestsTotal, metricsRequestsCost)
}

// RecordMetrics 记录指标
// 请注意需要先调用 InitMetrics
func RecordMetrics(serverName string, method, uri string, httpCode, businessCode int, costSeconds float64, traceId string) {
	// 不要在指标中出现 traceId ，容易导致指标爆炸
	_ = traceId
	metricsRequestsTotal.With(prometheus.Labels{
		"server_name":   serverName,
		"method":        method,
		"path":          uri,
		"http_code":     cast.ToString(httpCode),
		"business_code": cast.ToString(businessCode),
	}).Inc()

	metricsRequestsCost.With(prometheus.Labels{
		"server_name": serverName,
		"method":      method,
		"path":        uri,
	}).Observe(costSeconds)
}
