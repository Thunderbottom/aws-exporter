package exporter

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/sirupsen/logrus"
	"github.com/VictoriaMetrics/metrics"
)

// EC2Instance is a structure representing the functions required
// to fetch data from EC2
type EC2Instance struct {
	client  *ec2.EC2
	logger  *logrus.Logger
	metrics *metrics.Set
	region  string
}

// CollectInstanceMetrics scrapes the AWS EC2 API for Instance details and writes the metric data to Prometheus
func (exporter *Exporter) CollectInstanceMetrics() (error) {
	var client *ec2.EC2
	if exporter.Config.AWS.RoleARN != "" {
		creds := stscreds.NewCredentials(exporter.Session, exporter.Config.AWS.RoleARN)
		client = ec2.New(exporter.Session, &aws.Config{Credentials: creds})
	} else {
		client = ec2.New(exporter.Session)
	}

	ec2i := &EC2Instance{
		client:  client,
		logger:  exporter.Logger,
		metrics: exporter.Metrics,
		region:  exporter.Config.AWS.Region,
	}

	if err := ec2i.getInstanceUsage(); err != nil {
		return err
	}

	return nil
}

func (ec2i *EC2Instance) getInstanceUsage() (error) {
	var totalCoreCount int64

	instances, err := ec2i.client.DescribeInstances((&ec2.DescribeInstancesInput{}))
	if err != nil {
		ec2i.logger.Errorf("Error occurred while retreiving instance data: %s", err)
		return err
	}

	for _, reservation := range instances.Reservations {
		for _, instance := range reservation.Instances {
			if *instance.State.Name == "running" {
				totalCoreCount = totalCoreCount + *instance.CpuOptions.CoreCount
			}
			instanceTypeMetric := fmt.Sprintf(`ce_instance_count{region="%s",type="%s",status="%s"}`, ec2i.region, *instance.InstanceType, *instance.State.Name)
			ec2i.metrics.GetOrCreateCounter(instanceTypeMetric).Add(1)
		}
	}

	totalCoreMetric := fmt.Sprintf(`ce_total_core_count{region="%s"}`, ec2i.region)
	ec2i.metrics.GetOrCreateGauge(totalCoreMetric, func() float64 {
		return float64(totalCoreCount)
	})

	return nil
}
