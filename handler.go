package main

import (
	"context"
	"net/http"
	"sync"

	"golang.org/x/sync/errgroup"
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

			awsExporter.Job = &job
			awsExporter.Logger = Logger
			awsExporter.SetAWSSession()
			awsExporter.Metrics = metricSet
			var g errgroup.Group
			g.Go(awsExporter.CollectCostMetrics)
			g.Go(awsExporter.CollectInstanceMetrics)

			if err := g.Wait(); err != nil {
				cancel()
			}
		}()
	}
	wg.Wait()

	if ctx.Err() != nil {
		w.Write([]byte("An error has occurred, check logs for more information."))
	} else {
		metricSet.WritePrometheus(w)
	}
}
