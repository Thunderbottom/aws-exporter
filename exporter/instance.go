package exporter

import (
	"fmt"

	"github.com/VictoriaMetrics/metrics"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

// EC2Instance is a structure representing the functions required
// to fetch data from EC2
type EC2Instance struct {
	client  *ec2.EC2
	filters []*ec2.Filter
	job     string
	logger  *logrus.Logger
	metrics *metrics.Set
	region  string
}

// CollectInstanceMetrics scrapes the AWS EC2 API for Instance details and writes the metric data to Prometheus
func (exporter *Exporter) CollectInstanceMetrics() error {
	ec2i := exporter.getEC2Exporter()

	var g errgroup.Group
	g.Go(ec2i.getInstanceUsage)

	if err := g.Wait(); err != nil {
		return err
	}

	return nil
}
func (ec2i *EC2Instance) getInstanceUsage() error {
	var totalCoreCount int64

	input := &ec2.DescribeInstancesInput{}

	if len(ec2i.filters) != 0 {
		input.Filters = ec2i.filters
	}

	instances, err := ec2i.client.DescribeInstances(input)
	if err != nil {
		ec2i.logger.Errorf("Error occurred while retreiving instance data: %s", err)
		return err
	}

	for _, reservation := range instances.Reservations {
		for _, instance := range reservation.Instances {
			if *instance.State.Name == "running" {
				totalCoreCount = totalCoreCount + *instance.CpuOptions.CoreCount
			}
			instanceTypeMetric := fmt.Sprintf(`ce_instance_count{job="%s",region="%s",type="%s",status="%s"}`, ec2i.job, ec2i.region, *instance.InstanceType, *instance.State.Name)
			ec2i.metrics.GetOrCreateCounter(instanceTypeMetric).Add(1)
		}
	}

	totalCoreMetric := fmt.Sprintf(`ce_total_core_count{job="%s",region="%s"}`, ec2i.job, ec2i.region)
	ec2i.metrics.GetOrCreateGauge(totalCoreMetric, func() float64 {
		return float64(totalCoreCount)
	})

	return nil
}

func (exporter *Exporter) getEC2Exporter() *EC2Instance {
	var client *ec2.EC2
	if exporter.Job.AWS.RoleARN != "" {
		creds := stscreds.NewCredentials(exporter.Session, exporter.Job.AWS.RoleARN)
		client = ec2.New(exporter.Session, &aws.Config{Credentials: creds})
	} else {
		client = ec2.New(exporter.Session)
	}

	filters := make([]*ec2.Filter, 0, len(exporter.Job.Filters))
	for _, tag := range exporter.Job.Filters {
		if tag.Name != "" || tag.Value != "" {
			filters = append(filters, &ec2.Filter{
				Name:   aws.String(tag.Name),
				Values: []*string{aws.String(tag.Value)},
			})
		}
	}

	ec2i := &EC2Instance{
		client:  client,
		filters: filters,
		job:     exporter.Job.Name,
		logger:  exporter.Logger,
		metrics: exporter.Metrics,
		region:  exporter.Job.AWS.Region,
	}

	return ec2i
}
