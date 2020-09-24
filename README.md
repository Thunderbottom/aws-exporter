# Prometheus Exporter for AWS

## Overview

Export metrics from AWS Cost Explorer and EC2 Instance to Prometheus, using the AWS Go SDK.

## Getting Started

To get started, copy `config.toml.example` to `config.toml`. If you have `awscli` configured on your system, the AWS Go SDK will automatically fetch the credentials from your environment. If you would like to use other credentials, you'll need to set the `access_key` and `secret_key` inside `config.toml`.

If you want to scrape data from across AWS Accounts, you will also need to set `role_arn` to the IAM Role ARN of the assumed role.

The `region` needs to be set to the AWS region for which the data needs to be fetched.

### Required IAM Permissions

For the exporter to work, your IAM User/Role needs to have the following IAM Permissions attached:

```json
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Sid": "VisualEditor0",
            "Effect": "Allow",
            "Action": [
              "ce:GetCostForecast",
              "ce:GetCostAndUsage",
              "ce:GetReservationCoverage",
              "ce:GetReservationUtilization",
              "ec2:DescribeInstances"
            ],
            "Resource": "*"
        }
    ]
}
```

### Configuration

`aws-exporter` supports exporting data from multiple AWS accounts. For this, you need to create an entry for the account inside `config.toml`:

```toml
[[jobs]]
name = ""
[jobs.aws]
access_key = ""
secret_key = ""
region     = "ap-south-1"
role_arn   = ""
[jobs.filters]
name  = ""
value = ""
[jobs.exported_tags]
tag          = "ec2-tagname"
exported_tag = "ec2_tagname"
```

`[jobs.aws]` holds the credentials for the AWS account, and can be added per job. If no `access_key` and `secret_key` is specified, the exporter uses the default credentials configured by `awscli`.

**(Optional)** `[jobs.filters]` contains the filter tags to be applied for fetching EC2 Instance details. The `name` of the tag needs to be in the format `tag:tag-name`.

**(Optional)** `[jobs.exported_tags]` contains EC2 Instance Tags to search for (`tag`), and its corresponding tag name to be exported in the metric (`exported_tag`).

## Installation

### Compiling the Binary

```shell
$ git clone git@github.com:thunderbottom/aws-exporter.git
$ cd aws-exporter
$ make dist
$ cp config.toml.example config.toml
$ ./aws-exporter
```

### Docker Installation

To locally build and run the docker image, make sure you have edited `config.toml` before running:

```shell
$ docker build -t aws-exporter -f docker/Dockerfile .
$ docker run -p 9980:9980 -v config.toml:/config.toml aws-exporter
```

If you do not want to build your own docker image:

```shell
$ docker run -p 9980:9980 -v config.toml:/config.toml thunderbottom/aws-exporter
```

## Advanced

### Setting Up Prometheus

Add the following configuration to Prometheus:

```yaml
- job_name: 'aws-exporter'
  metrics_path: '/metrics'
  static_configs:
  - targets: ['localhost:9980']
    labels:
      service: aws-exporter
```

### Adding more exporters

Extending the functionality of this exporter is easy. Add the exporter under the `exporter` package, implement the `*Exporter` structure for your metrics collector, and create a structure for the exporter functions. Register your exporter inside `handler.go`. You may take a look at the existing exporters to get an idea.

## Contributions

PRs for feature requests, bug fixes are welcome. Feel free to open an issue for bugs and discussions on the exporter functionality.

## License

```
MIT License

Copyright (c) 2020 Chinmay Pai

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
```
