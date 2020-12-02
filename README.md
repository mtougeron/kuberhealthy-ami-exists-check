# kuberhealthy-ami-exists-check
A Kuberhealthy check to make sure the AWS AMIs in use still exist.

![Go](https://github.com/mtougeron/kuberhealthy-ami-exists-check/workflows/Go/badge.svg) ![Gosec](https://github.com/mtougeron/kuberhealthy-ami-exists-check/workflows/Gosec/badge.svg) ![ContainerScan](https://github.com/mtougeron/kuberhealthy-get-pods-check/workflows/ContainerScan/badge.svg) [![GitHub tag](https://img.shields.io/github/tag/mtougeron/kuberhealthy-ami-exists-check.svg)](https://github.com/mtougeron/kuberhealthy-ami-exists-check/tags/)

The `Kuberhealthy Get AMI Check` gathers the list Nodes, finds the AWS Instances, creates a unique list of AMI IDs, and verifies that the AMI(s) still exist.

## Thanks Comcast!

A big shout-out and thank you goes to Comcast for writing [Kuberhealthy](https://github.com/Comcast/kuberhealthy)

A large part of this check is due to the preceding work done https://github.com/Comcast/kuberhealthy/tree/master/cmd/ami-check

## Kuberhealthy AMI Exists Spec Example

```yaml
apiVersion: comcast.github.io/v1
kind: KuberhealthyCheck
metadata:
  name: ami-exists
spec:
  runInterval: 30m
  timeout: 1m
  extraAnnotations:
    iam.amazonaws.com/role: <role-arn> # Replace this value with your ARN
    iam.amazonaws.com/external-id: <role-external-id> # Use this if kube2iam is using external-id for roles
  podSpec:
    serviceAccountName: khcheck-ami-exists
    containers:
    - name: main
      image: ghcr.io/mtougeron/khcheck-ami-exists:latest # Change to a specific version
      imagePullPolicy: IfNotPresent
      env:
        - name: DEBUG
          value: "1"
        - name: AWS_REGION
          value: "<region name>" # The region your cluster runs in
```

### Installation

> Make sure you are using the latest release of Kuberhealthy 2.x

Create the AWS Role that will be assumed so that the check has access to `ec2:DescribeInstances` & `ec2:DescribeImages` using the [example role policy](example/aws-role.json).

Run `kubectl apply` against [example spec file](example/khcheck-ami-exists.yaml). NOTE: This also installs the `ClusterRole` and `RoleBinding` to allow the check to get the list of `Nodes` running in the cluster.

```bash
kubectl apply -f khcheck-ami-exists.yaml -n kuberhealthy
```

#### Container Image

Images are available on the [GitHub Container Registry](https://github.com/users/mtougeron/packages/container/khcheck-ami-exists/versions) and [DockerHub](https://hub.docker.com/repository/docker/mtougeron/khcheck-ami-exists)

### Licensing

This project is licensed under the Apache V2 License. See [LICENSE](LICENSE) for more information.
