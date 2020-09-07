// Licensed to Michael Tougeron <github@e.tougeron.com> under
// one or more contributor license agreements. See the LICENSE
// file distributed with this work for additional information
// regarding copyright ownership.
// Michael Tougeron <github@e.tougeron.com> licenses this file
// to you under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

// NOTE: This check leverages heavily the work done in the Kuberhealthy
// ami-check https://github.com/Comcast/kuberhealthy/tree/master/cmd/ami-check

package main

import (
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
	log "github.com/sirupsen/logrus"
)

// AMIResult struct represents a query for AWS AMIs. Contains a list
// of AMIs and an error.
type AMIResult struct {
	Images []*ec2.Image
	Err    error
}

// InstanceAMIsResult is a list of the AMI IDs
type InstanceAMIsResult struct {
	InstanceAMIs []*string
	Err          error
}

// Client EC2 client interface
type Client struct {
	ec2iface.EC2API
}

func createAWSSession() *session.Session {
	// Build an AWS session
	log.Debugln("Building AWS session")
	awsConfig := aws.NewConfig().WithCredentialsChainVerboseErrors(debug)
	awsConfig.Region = aws.String(awsRegion)
	minThrottleDelay, _ := time.ParseDuration("200ms")
	maxThrottleDelay, _ := time.ParseDuration("30s")
	awsConfig.Retryer = CustomRetryer{DefaultRetryer: client.DefaultRetryer{
		NumMaxRetries:    5,
		MinThrottleDelay: minThrottleDelay,
		MaxThrottleDelay: maxThrottleDelay,
	}}

	return session.Must(session.NewSession(awsConfig))
}

// EC2Client initializes an EC2 client
func newEC2Client() (*Client, error) {
	svc := ec2.New(awsSession, &aws.Config{Region: aws.String(awsRegion)})
	return &Client{svc}, nil
}

// listEC2Images gets the specific list of AMIs based on their IDs
func (c *Client) listEC2Images(imageIDs []*string) chan AMIResult {

	listChan := make(chan AMIResult)

	go func() {
		defer close(listChan)

		amiResult := AMIResult{}

		images, err := c.DescribeImages(&ec2.DescribeImagesInput{
			ImageIds: imageIDs,
		})

		if err != nil {
			amiResult.Err = err
			listChan <- amiResult
			return
		}

		amiResult.Images = images.Images
		listChan <- amiResult
		return
	}()

	return listChan
}

// listEC2InstanceAMIs collects the AMI IDs used by the Instances
func (c *Client) listEC2InstanceAMIs(instanceIDs []*string) chan InstanceAMIsResult {

	listChan := make(chan InstanceAMIsResult)

	go func() {
		defer close(listChan)

		InstanceAMIsResult := InstanceAMIsResult{}

		foundAMIs := map[string]bool{}

		err := c.DescribeInstancesPages(&ec2.DescribeInstancesInput{InstanceIds: instanceIDs},
			func(page *ec2.DescribeInstancesOutput, lastPage bool) bool {
				for _, reservation := range page.Reservations {
					for _, instance := range reservation.Instances {
						if !foundAMIs[*instance.ImageId] {
							InstanceAMIsResult.InstanceAMIs = append(InstanceAMIsResult.InstanceAMIs, instance.ImageId)
							foundAMIs[*instance.ImageId] = true
						}
					}
				}
				return !lastPage
			})

		if err != nil {
			InstanceAMIsResult.Err = err
			listChan <- InstanceAMIsResult
			return
		}

		listChan <- InstanceAMIsResult
		return
	}()

	return listChan
}
