package main

import (
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/thunderbottom/aws-exporter/config"
)

var (
	// Cfg is an instance of Config containing the app configuration
	Cfg         = config.GetConfig()
	// Logger is an instance of logrus.Logger to be used throughout the exporter
	Logger      = getLogger()
)

func getLogger() *logrus.Logger {
	logger := logrus.New()
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})
	logger.SetLevel(logrus.InfoLevel)

	return logger
}

func main() {
	router := http.NewServeMux()
	router.Handle("/", http.HandlerFunc(defaultHandler))
	router.Handle("/metrics", http.HandlerFunc(metricsHandler))

	server := &http.Server{
		Addr:         Cfg.Server.Address,
		Handler:      router,
		ReadTimeout:  Cfg.Server.ReadTimeout * time.Millisecond,
		WriteTimeout: Cfg.Server.WriteTimeout * time.Millisecond,
	}

	Logger.Infof("Starting server. Listening on: %v", Cfg.Server.Address)
	if err := server.ListenAndServe(); err != nil {
		Logger.Fatalf("error starting server: %v", err)
	}
}
