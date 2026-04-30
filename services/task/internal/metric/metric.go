package metric

import "github.com/prometheus/client_golang/prometheus"

var (
	TaskCreatedTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "task_created_total",
		Help: "创建任务总数",
	})

	TaskCompletedTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "task_completed_total",
		Help: "完成任务总数",
	})

	TaskDuration = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "task_duration_seconds",
		Help:    "任务处理耗时",
		Buckets: prometheus.DefBuckets,
	})
)

func Init() {
	prometheus.MustRegister(TaskCreatedTotal)
	prometheus.MustRegister(TaskCompletedTotal)
	prometheus.MustRegister(TaskDuration)
}
