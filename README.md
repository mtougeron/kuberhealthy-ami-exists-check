# kuberhealthy-get-ami-check
A Kuberhealthy check to make sure the AWS AMIs in use still exist

![Go](https://github.com/mtougeron/kuberhealthy-get-ami-check/workflows/Go/badge.svg) ![Gosec](https://github.com/mtougeron/kuberhealthy-get-ami-check/workflows/Gosec/badge.svg) [![GitHub tag](https://img.shields.io/github/tag/mtougeron/kuberhealthy-get-ami-check.svg)](https://github.com/mtougeron/kuberhealthy-get-ami-check/tags/)

The `Kuberhealthy Get AMI Check` gathers the list of AWS Instances running, gets their AMI IDs, and verifies that the AMIs still exist.

## Thanks Comcast!

A big shout-out and thank you goes to Comcast for writing [Kuberhealthy](https://github.com/Comcast/kuberhealthy)

A large part of this check is due to the preceding work done https://github.com/Comcast/kuberhealthy/tree/master/cmd/ami-check
