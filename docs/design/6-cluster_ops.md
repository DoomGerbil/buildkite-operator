# Running the Operator

When you deploy the Operator into a cluster, you can also provide some configuration options, if needed.
These live in a `ConfigMap` in the Operator's namespace creatively named `buildkite-operator-config`.

- `job_namespace` (optional, but highly recommended)
  - What namespace to run jobs in.
  - If unset, the Operator's namespace is used.
- `log_exporter_config`
  - TODO: TBD
- `metrics_exporter_config`
  - TODO: TBD
- TODO: Other stuff goes here

There are also a few advanced topics worth considering.

## Enable Cluster Node Autoscaling

Depending on how heavyweight your jobs are, and how cyclical the workload patterns are, it is likely
to be highly beenficial to enable node autoscaling for your cloud cluster in order to add capacity
when you need it, and shut it down when you don't.

This is out of scope of this document - see your cloud provider's docs for details:

[GKE autoscaling](https://cloud.google.com/kubernetes-engine/docs/concepts/cluster-autoscaler)
[EKS autoscaling](https://docs.aws.amazon.com/eks/latest/userguide/cluster-autoscaler.html)
[AKS autoscaling](https://docs.microsoft.com/en-us/azure/aks/cluster-autoscaler)

## Use Heterogenous Node Layouts

Some workloads may have different requirements that benefit from heterogenous nodes. For example:

- You may have some Linux workloads and some Windows workloads
- You might have specific jobs that require GPUs
- You might have some very large jobs that require dedicated nodes.

There are several ways to configure a `BuildJob` with these requirements.

### Job Requirements

If a job has a requirement, and can _only_ run on nodes that match, you can specify
`node_requirements`. These jobs will be guaranteed to only run on nodes that match, and will not be
able to execute until these nodes are available.

This maps to the Kubernetes `nodeAffinity` concept of `requiredDuringSchedulingIgnoredDuringExecution`.

TODO: Provide examples of this

### Job Preferences

If a job has a preference - for example, if a workload wants a local SSD for performance, but should still
run if no local SSDs are available - a `BuildJob` can specify `JobNodePreferences`. The job will
_prefer_ running on nodes that meet the specification, but will still be run elsewhere if the request
cannot be met.

This maps to the Kubernetes `nodeAffinity` concept of `preferredDuringSchedulingIgnoredDuringExecution`.

TODO: Provide examples of this

### Node Reservation (taints and tolerances)

In some cases, you may want to set aside some nodes for _only_ specific workloads to use. In this case, you can use the Kubernetes concepts of `taints` and `tolerations`, which marks nodes as generally off limits
except for specific workloads that explicitly choose to run on them.

TODO: Provide examples of this
