# Buildkite Operator Design

The overall goal of this project is to enable users to easily run BuildKite jobs in a scalable,
efficient manner using Kubernetes clusters to host workload execution.

Many of the design principles and implementation details have been heavily inspired by the Kubernetes
project's CI system, [Prow], and a lot of lessons have been learned as well.

## Core End-User Workflow

The core usage loop of the Operator is simple:

1. A user creates a `BuildJob` resource in the cluster.
1. The controller creates a pod - the configuration of which is described in [`BuildEnvironment`](#BuildEnvironment)) - based on the supplied spec.
1. The job runs inside this container, eventually exiting and reporting a status code.
    - This runs the standard buildkite-agent workflow, starting with `buildkite-agent bootstrap`.
    - Thus, log streaming and other standard `buildkite-agent` functionality should Just Work.
1. When the workload has completed, the controller will reap the completed resource immediately.

### Caveats

Since users are responsible for creating `BuildJob` resources in the cluster
for every BuildKite job, there are two implicit dependencies:

- There must be _at least one_ agent running outside of the context of the Operator
  - This agent would be responsible for actually inserting the `BuildJob` resource into the cluster for _every_ job.
  - And thus must have cluster access permissions
  - And would be a single point of failure
- The BuildKite step definition has to be converted into a `BuildJob` resource for the Operator to schedule it.
  - Thus, every job must either:
    - Have a pre-created static `BuildJob` definition stored somewhere
    - Generate a dynamic `BuildJob` definition at runtime
    - Use a BuildKite Plugin to generate a `BuildJob` definition
  - And then the non-Operator agent would create the resource in the cluster.

See the [JobSpawner section](#JobSpawner) below for the proposed solution to this issue.

## BuildJobs

A `BuildJob` is the basic resource used to represent the job to be run. Creating this resource in the
cluster is what initiates actual workload execution, and its lifecycle follows the workload it represents.

### The Life of a BuildJob

- The `BuildJob` resource is created in the cluster.
- The controller reads the spec and creates a workload for it. This workload pod will contain:
  - Zero or more `initContainers`
    - These may do any sort of initialization prework required, and are guaranteed to complete before the job starts.
      - For example, using [repo mirroring](#RepoMirror) will create an `initContainer` that mounts the requested repo mirrors to the pod.
    - `initContainers` run in serial, and block workload startup, so adding lots of these can have a negative performance impact.
    - All `initContainer` failures are considered fatal, and mean that the primary container will never be started.
  - The primary workload container
    - In many cases, this can be a reference directly to a published docker container
      - eg `golang:1.13` or `node:14.8.0-alpine3.11`
    - For customized container environments, see [BuildEnvironments](#BuildEnvironments)
  - One or more sidecar containers
    - These run alongside the primary workload container and mount a shared filesystem.
    - The `Buildkite Agent` sidecar container will always be injected into every pod.
    - Other provided sidecars are documented below
    - Custom sidecars can also be specified - see [BuildEnvironments](#BuildEnvironments) for how to create these.

### BuildJob Configuration Options

See the [BuildJob spec](../../api/v1alpha1/buildjob_types.go) for configuration options.

TODO actual docs on options, examples, etc

## BuildEnvironments

TODO: rethink this name. Maybe BuildEnvironment is more than just a single container, and we need a
  lower-level concept that is encapsulated within this.

This resource defines a container environment. This may be the environment in which a job will run, or
it may be a supplimental container that's used by other `BuildEnvironments` as a sidecar or an initContainer.

### Init Containers

A set of containers that run before the main workload, to ensure that pre-work is done before the job starts.

#### RepoMirror

This container will check the cluster-local repo cache to see if the requested git repo is present, and
will copy it into the pod if so. It will also ensure that the pod's clone of the repo is up to date before

See the [RepoSyncer](#RepoSyncer) section for details on how these repo mirrors are kept up to date.

### Sidecar Containers

A set of sidecar containers will mounted into the pod alongside the primary workload container.

If you don't want the added functionality, most of these can be disabled in the `BuildJob` resource spec - only the `BuildKite Agent` sidecar is required.

#### BuildKite Agent Sidecar

This contains the `buildkite-agent` binary at a specified version, which will be mounted into the pod.

TODO: where? `/buildkite-agent`? On the PATH somewhere? `/usr/bin`?

It is also responsible for ensuring that the pod is authenticated to buildkite.com so that it can be
used for artifacts, annotations, meta-data, etc.

TODO: Does this also run the `buildkite-agent bootstrap`? If not, where do we do that, since we want to hook into that somewhere?

#### LogExporter Sidecar

This container can be attached to a job along with a per-job configuration, which can export logs three ways:

1. The job can supply a list of filepaths,
1. The job can mount a volume into the pod, and any files written there will be captured as logs,
1. The job can specify to capture stdout/stderr.

This also has some cluster-level configuration for setting up log delivery mechanisms - eg how to
send logs to ElasticSearch, etc.

#### ArtifactExporter

This container will create a volume in the pod, and any files written to it will be captured as
a Buildkite artifact.

#### MetricsExporter

This container will export metrics from the job, from any file written to a specified path. Export may
happen via either:

1. Pushing metrics directly to a pushgateway
    - Pushgateway must be set up in the cluster
    - The pushgateway location must be configured for the cluster
1. Exporting metrics on a known port
    - Prometheus must be configured to monitor the service (out of scope of this doc)

#### Custom Sidecar Containers

You can define a custom sidecar container the same way you define a `BuildEnvironment`. Once defined,
just specify is as an `extraContainer` in your `BuildJob` spec and it will be attached to your pod.

For example, maybe you want to have a test result exporter that watches for the main workload to complete,
and then collects a JUnit test report XML output and sends it to an external service for processing/reporting.

By putting this in a sidecar instead of in your primary workload container, it becomes both reusable across
multiple jobs, and allows you to make changes to your test reporting without needing to change your build
environment.

## A Word About Repo Mirroring

Standard persistent disks in a cloud K8S environment cannot be mounted as writable by multiple pods at
once - they can be used `ReadOnceMany`, but not `ReadWriteMany`, which means that there can only be
one "owner" of the local mirror responsible for keeping the mirror updated.

In this system, that is the `reposyncer` service, which mounts the repo mirror disk as read-write,
where all other pods using it mount the disk read-only and clone the remote repo as a shared reference
to the mirror like so:

```shell
git clone \
  git@github.com:kubernetes/kubernetes.git \
  --quiet \
  --shared \
  --reference /path/to/mounted/repo/mirror \
  /path/to/local/clone/target
```

This set of flags will ensure the minimum amount of traffic transferred to the instance:

- `--shared` - The clone will start without _any_ local objects copied from the `reference`
- `--reference` - The clone will pull everything possible from the local reference repo instead of
the remote.

## Controller

The controller is largely responsible for ingesting the resource, creating the pod, and then handing off
control to the `buildkite-agent` bootstrap hook in the container.

After that, the controller's responsibility is mostly to ensure that the container gets cleaned up after
it's complete or timed out.

## BuildJobValidatingWebhook

This webhook will reject any invalid job resource at creation time, ensuring that a job with a spec
that is impossible to schedule will be rejected.

However, some jobs that are invalid may not always be detectable.

TODO (future): Provide some way to allow custom, user-defined validation.

## BuildEnvironmentMutatingWebhook

This webhook is responsible for automatically modifying a created `BuildEnvironment` to attach sidecars
and setting any other necessary configuration required.

TODO: Is this necessary in particular? Have to figure it out.

## JobSpawner

As discussed [above](#Caveats), having to run a separate, static agent outside of the Operator is
fine for a proof of concept, but overall adds complexity that partially defeats the purpose of using
an Operator to orchestrate builds.

This service registers with BuildKite as an agent, and when assigned a job by
BuildKite, will create `BuildJob` resources in the cluster based on the step data.

This serves a few purposes:

- Users no longer need to generate and create `BuildJob` K8S resources themselves.
  - So using a Buildkite Plugin is no longer necessary.
- You
- Users running a job no longer need to have direct cluster access

Thus, the end-user workflow changes to instead be:

1. A BuildKite pipeline runs, and creates one or more BuildKite jobs targeting the `JobSpawner`.
1. The `JobSpawner` creates `BuildJob` resources in the cluster based on the settings specified.

And the rest of the workflow proceeds as before.

### Targeting BuildKite Jobs to run on the JobSpawner

The `JobSpawner` registers as a BuildKite agent, and as such, standard [BuildKite agent targeting rules] apply.

By default, `JobSpawner` will register with a single label - `queue=bk-operator` - and can thus be targeted like so:

```yaml
steps:
  - command: "build.sh"
    agents:
      queue: bk-operator
```

If extra labels are specified in the `JobSpawner`'s configuration, it will register with those as well.
This can be useful if, for example, you have multiple Operators (eg in different clusters) at which
you wish to target different jobs, like so:

```yaml
steps:
  - command: "deploy-gke.sh"
    agents:
      queue: bk-operator
      cluster: gke-prod
  - command: "deploy-eks.sh"
    agents:
      queue: bk-operator
      cluster: eks-prod
```

## RepoSyncer

In order to speed up repo syncing, primary workload containers will mount the mirror disk read-only
and reference the mirror when cloning.

This component is responsible for keeping those persistent repo mirrors up to date. Ordinarily, we could
let `buildkite-agent` keep these mirrors up to date, but in this case, we want to have a persistent
mirror shared between all ephemeral containers, or else there's no point in mirroring anyway.

Those mirrors need to be kept close-ish in time to the remote, and that's what this service does, so
that any fetches needed by containers are small deltas.

## Running the Operator

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

### Enable Cluster Autoscaling

Depending on how heavyweight your jobs are, and how cyclical the workload patterns are, it is likely
to be highly beenficial to enable node autoscaling for your cloud cluster in order to add capacity
when you need it, and shut it down when you don't.

This is out of scope of this document - see your cloud provider's docs for details:

[GKE autoscaling](https://cloud.google.com/kubernetes-engine/docs/concepts/cluster-autoscaler)
[EKS autoscaling](https://docs.aws.amazon.com/eks/latest/userguide/cluster-autoscaler.html)
[AKS autoscaling](https://docs.microsoft.com/en-us/azure/aks/cluster-autoscaler)

### Heterogenous Node Layouts

Some workloads may have different requirements that benefit from heterogenous nodes. For example:

- You may have some Linux workloads and some Windows workloads
- You might have specific jobs that require GPUs
- You might have some very large jobs that require dedicated nodes.

There are several ways to configure a `BuildJob` with these requirements.

#### Job Requirements

If a job has a requirement, and can _only_ run on nodes that match, you can specify
`JobNodeRequirements`. These jobs will be guaranteed to only run on nodes that match, and will not be
able to execute until these nodes are available.

This maps to the Kubernetes `nodeAffinity` concept of `requiredDuringSchedulingIgnoredDuringExecution`.

#### Job Preferences

If a job has a preference - for example, if a workload wants a local SSD for performance, but should still
run if no local SSDs are available - a `BuildJob` can specify `JobNodePreferences`. The job will
_prefer_ running on nodes that meet the specification, but will still be run elsewhere if the request
cannot be met.

This maps to the Kubernetes `nodeAffinity` concept of `preferredDuringSchedulingIgnoredDuringExecution`.

#### Node Reservation (taints and tolerances)

In some cases, you may want to set aside some nodes for _only_ specific workloads to use. In this case, you can use the Kubernetes concepts of `taints` and `tolerations`, which marks nodes as generally off limits
except for specific workloads that explicitly choose to run on them.

TODO TBD and docs

[Prow]: https://github.com/kubernetes/test-infra/tree/master/prow
[BuildKite agent targeting rules]: https://buildkite.com/docs/agent/v3/cli-start#agent-targeting
