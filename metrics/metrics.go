package metrics

import "github.com/prometheus/client_golang/prometheus"

var (
	jobLabelNames = []string{"name"}
	JobsCreated   = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "docker",
		Subsystem: "jobs",
		Name:      "created_total",
		Help:      "Number of jobs created.",
	}, jobLabelNames)
	JobsRan = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "docker",
		Subsystem: "jobs",
		Name:      "ran_total",
		Help:      "Number of jobs ran.",
	}, jobLabelNames)
	JobsRunFailed = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "docker",
		Subsystem: "jobs",
		Name:      "run_failed_total",
		Help:      "Number of jobs failed during run.",
	}, jobLabelNames)
	JobLatency = prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Namespace: "docker",
		Subsystem: "jobs",
		Name:      "latency",
		Help:      "Latency of job created.",
	}, jobLabelNames)
)

func init() {
	prometheus.MustRegister(JobsCreated)
	prometheus.MustRegister(JobsRan)
	prometheus.MustRegister(JobsRunFailed)
	prometheus.MustRegister(JobLatency)
}
