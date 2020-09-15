package main

import (
	"net/http"

	"github.com/thunderbottom/aws-exporter/exporter"
	"github.com/VictoriaMetrics/metrics"
)

var (
	awsExporter = &exporter.Exporter{}
)

func defaultHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Welcome to prometheus cost exporter. Visit /metrics."))
}

func metricsHandler(w http.ResponseWriter, r *http.Request) {
	metricSet := metrics.NewSet()

	for _, job := range Cfg.Jobs {
		j := job
		awsExporter.Job = &j
		awsExporter.Logger = Logger
		awsExporter.SetAWSSession()
		awsExporter.Metrics = metricSet
		if err := awsExporter.CollectCostMetrics(); err != nil {
			w.Write([]byte("An error has occurred while collecting Cost Exporter metrics, check the logs for more information."))
			return
		}
		if err := awsExporter.CollectInstanceMetrics(); err != nil {
			w.Write([]byte("An error has occurred while collecting EC2 Instance metrics, check the logs for more information."))
			return
		}
	}

	metricSet.WritePrometheus(w)
}
