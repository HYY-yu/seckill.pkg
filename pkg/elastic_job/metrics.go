package elastic_job

import (
	"github.com/prometheus/client_golang/prometheus"
)

type JobMetrics struct {
	namespace string

	metricsJobAddTotal *prometheus.CounterVec
	metricsJobRunCost  *prometheus.HistogramVec
}

func NewJobMetrics(namespace string) *JobMetrics {
	j := new(JobMetrics)

	j.metricsJobAddTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "metrics_elastic_job_add_total",
			Help:      " The total number of calls to AddJob during the program run. ",
		},
		[]string{"server_name", "job_key", "trace_id"},
	)

	j.metricsJobRunCost = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "metrics_elastic_job_run_cost_seconds",
			Help:      " Job running time.  ",
			Buckets:   []float64{.005, .01, .025, .05, .1, 1, 2.5, 5},
		},
		[]string{"server_name", "job_key"},
	)

	// 自动注册到 prometheus Default
	prometheus.MustRegister(j.metricsJobAddTotal, j.metricsJobRunCost)
	return j
}

func (j *JobMetrics) MetricsAddTotal(serverName, jobKey, traceId string) {
	if len(serverName) == 0 {
		serverName = "metrics_job"
	}

	if len(traceId) == 0 {
		traceId = "none"
	}

	j.metricsJobAddTotal.With(
		prometheus.Labels{
			"server_name": serverName,
			"job_key":     jobKey,
			"trace_id":    traceId,
		}).Inc()
}

func (j *JobMetrics) MetricsRunCost(serverName, jobKey string, costSeconds float64) {
	if len(serverName) == 0 {
		serverName = "metrics_job"
	}

	j.metricsJobRunCost.With(
		prometheus.Labels{
			"server_name": serverName,
			"job_key":     jobKey,
		}).Observe(costSeconds)
}
