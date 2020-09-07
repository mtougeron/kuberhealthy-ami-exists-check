package main

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
	"github.com/stretchr/testify/assert"
)

// Define a mock struct to be used in unit tests
type mockEC2Client struct {
	ec2iface.EC2API
	resp   ec2.DescribeInstancesOutput
	result InstanceAMIsResult
}

func (m *mockEC2Client) DescribeInstancesPages(*ec2.DescribeInstancesInput, func(*ec2.DescribeInstancesOutput, bool) bool) error {
	return nil
}

func Test_listEC2InstanceAMIs(t *testing.T) {
	results := InstanceAMIsResult{}
	results.InstanceAMIs = append(results.InstanceAMIs, aws.String("ami-123abc"))
	results.Err = nil
	cases := []mockEC2Client{
		{
			resp: ec2.DescribeInstancesOutput{
				Reservations: []*ec2.Reservation{
					{
						Instances: []*ec2.Instance{
							{
								ImageId: aws.String("ami-123abc"),
							},
							{
								ImageId: aws.String("ami-123abc"),
							},
							{
								ImageId: aws.String("ami-123abc"),
							},
						},
					},
				},
			},
			result: results,
		},
	}

	var instanceIDs []*string
	instanceIDs = append(instanceIDs, aws.String("i-abc123"))
	instanceIDs = append(instanceIDs, aws.String("i-abc456"))
	instanceIDs = append(instanceIDs, aws.String("i-abc789"))
	for _, c := range cases {
		e := Client{
			&mockEC2Client{
				resp:   c.resp,
				result: c.result,
			},
		}

		var instanceResult InstanceAMIsResult
		instanceResult = <-e.listEC2InstanceAMIs(instanceIDs)

		// amiIDs := e.listEC2InstanceAMIs(instanceIDs)

		assert := assert.New(t)

		assert.EqualValues(c.result, instanceResult)
	}
}
