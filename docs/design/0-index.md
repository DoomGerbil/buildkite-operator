# Buildkite Operator

The overall goal of this project is to enable users to easily run BuildKite jobs in a scalable,
efficient manner using Kubernetes clusters to host workload execution.

Many of the design principles and implementation details have been heavily inspired by the Kubernetes
project's CI system, [Prow], and a lot of lessons have been learned as well.

## End User Walkthrough

If you're an end user who just wants to know how to use this for builds, [start here](1-user_guide.md).

## Life of a BuildJob

To understand the entire lifecycle of a `BuildJob`, [see here](2-build_job.md).

### BuildJob Configuration API

## BuildEnvironment

To understand when and how to define a `BuildEnvironment`, [see here](3-build_environment.md).

### Pod Decorators, Init Containers, and Sidecars

### Provided Containers

#### buildkite-agent Sidecar

#### log-exporter Sidecar

#### artifact-exporter Sidecar

#### annotation-exporter Sidecar

#### metrics-exporter Sidecar

#### Custom Sidecars

## Repo Mirroring

Repo Mirroring is an especially complex topic - [see here](4-repo_mirroring.md) for details.

## Controllers and Services

For a breakdown of the different controllers and services that make up this Operator, [see here](5-controllers-services.md).

### Bob the BuildJob Builder

Name TBD

### Job Spawner

#### Targeting Rules

### Repo Syncer

### Validating and Mutating Webhooks

## Cluster Operations and Tuning

For advanced topics around cluster management and operations, include discussion of cloud provider platforms, [see here](6-cluster_ops.md).

### Jobs and Namespaces

### RBAC

### Cluster Autoscaling

### Multiple Cluster Operation

### Heterogenous Node Layouts

### Node Affinities, Taints, and Tolerances

### Differences in Cluster Providers

[Prow]: https://github.com/kubernetes/test-infra/tree/master/prow
