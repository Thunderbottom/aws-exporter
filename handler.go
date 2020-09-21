package main

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	"github.com/VictoriaMetrics/metrics"
	"github.com/thunderbottom/aws-exporter/exporter"
	"golang.org/x/sync/errgroup"
)

func defaultHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Welcome to prometheus cost exporter. Visit /metrics."))
}

func metricsHandler(w http.ResponseWriter, r *http.Request) {
	metricSet := metrics.NewSet()

	wg := sync.WaitGroup{}
	wg.Add(len(Cfg.Jobs))
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for _, job := range Cfg.Jobs {
		job := job
		go func() {
			defer wg.Done()
			select {
			case <-ctx.Done():
				return
			default:
			}

			awsExporter := &exporter.Exporter{}
			awsExporter.Job = &job
			awsExporter.Logger = Logger
			awsExporter.SetAWSSession()
			awsExporter.Metrics = metricSet
			var g errgroup.Group
			g.Go(awsExporter.CollectCostMetrics)
			g.Go(awsExporter.CollectInstanceMetrics)

			var status float64 = 1
			exporterUp := fmt.Sprintf(`ce_up{job="%s"}`, awsExporter.Job.Name)
			if err := g.Wait(); err != nil {
				cancel()
				awsExporter.Logger.Error(err)
				status = 0
			}
			awsExporter.Metrics.GetOrCreateGauge(exporterUp, func() float64 {
				return status
			})
		}()
	}
	wg.Wait()

	metricSet.WritePrometheus(w)

}
