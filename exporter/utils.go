package exporter

import (
	"github.com/sirupsen/logrus"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/VictoriaMetrics/metrics"
	"github.com/thunderbottom/aws-exporter/config"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
)

// Exporter representst the structure for all app wide exporters and structs
type Exporter struct {
	Job     *config.Job
	Logger  *logrus.Logger
	Session *session.Session
	Metrics *metrics.Set
}

// SetAWSSession is a method to create a new session for the AWS API
func (exporter *Exporter) SetAWSSession() {
	config := &aws.Config{
		Region: aws.String(exporter.Job.AWS.Region),
	}
	if exporter.Job.AWS.AccessKey != "" && exporter.Job.AWS.SecretKey != "" {
		config.Credentials = credentials.NewStaticCredentials(
				exporter.Job.AWS.AccessKey,
				exporter.Job.AWS.SecretKey,
				"")
	}
	exporter.Session = session.Must(session.NewSessionWithOptions(session.Options{
		Config: *config,
	}))
}
