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

type mockEC2ClientAMI struct {
	ec2iface.EC2API
	resp   ec2.DescribeImagesOutput
	result AMIResult
}

func (m *mockEC2Client) DescribeInstancesPages(in *ec2.DescribeInstancesInput, fn func(*ec2.DescribeInstancesOutput, bool) bool) error {
	fn(&m.resp, true)
	return nil
}

func (m *mockEC2ClientAMI) DescribeImages(*ec2.DescribeImagesInput) (*ec2.DescribeImagesOutput, error) {
	return &m.resp, nil
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

func Test_listEC2ImagesSuccess(t *testing.T) {
	results := AMIResult{}
	var img ec2.Image
	img.ImageId = aws.String("ami-123abc")
	results.Images = append(results.Images, &img)
	results.Err = nil
	cases := []mockEC2ClientAMI{
		{
			resp: ec2.DescribeImagesOutput{
				Images: []*ec2.Image{
					{
						ImageId: aws.String("ami-123abc"),
					},
				},
			},
			result: results,
		},
	}

	var imageIDs []*string
	imageIDs = append(imageIDs, aws.String("ami-123abc"))
	for _, c := range cases {
		e := Client{
			&mockEC2ClientAMI{
				resp:   c.resp,
				result: c.result,
			},
		}

		var imageResult AMIResult
		imageResult = <-e.listEC2Images(imageIDs)

		assert := assert.New(t)

		assert.Equal(c.result, imageResult)
	}
}

func Test_listEC2ImagesFail(t *testing.T) {
	results := AMIResult{}
	var img ec2.Image
	img.ImageId = aws.String("ami-123abc")
	results.Images = append(results.Images, &img)
	results.Err = nil
	cases := []mockEC2ClientAMI{
		{
			resp: ec2.DescribeImagesOutput{
				Images: []*ec2.Image{
					{
						ImageId: aws.String("ami-123abc"),
					},
					{
						ImageId: aws.String("ami-456def"),
					},
				},
			},
			result: results,
		},
	}

	var imageIDs []*string
	imageIDs = append(imageIDs, aws.String("ami-123abc"))
	for _, c := range cases {
		e := Client{
			&mockEC2ClientAMI{
				resp:   c.resp,
				result: c.result,
			},
		}

		var imageResult AMIResult
		imageResult = <-e.listEC2Images(imageIDs)

		assert := assert.New(t)

		assert.NotEqual(c.result, imageResult)
	}
}
