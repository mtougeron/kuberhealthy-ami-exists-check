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

func (m *mockEC2Client) DescribeInstancesPages(in *ec2.DescribeInstancesInput, fn func(*ec2.DescribeInstancesOutput, bool) bool) error {
	fn(&m.resp, true)
	return nil
}

func Test_listEC2InstanceAMIsSuccess(t *testing.T) {
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

		assert := assert.New(t)

		assert.Equal(c.result, instanceResult)
	}
}

func Test_listEC2InstanceAMIsFail(t *testing.T) {
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
								ImageId: aws.String("ami-456def"),
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

		assert := assert.New(t)

		assert.NotEqual(c.result, instanceResult)
	}
}
