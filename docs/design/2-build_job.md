# BuildJobs

A `BuildJob` is the Kubernetes resource used to represent the job to be run. Creating this resource in the
cluster initiates actual workload execution, and its lifecycle follows the workload it represents.

## The Life of a BuildJob

- The `BuildJob` resource is created in the cluster.
- The controller ingests the spec and creates a workload for it. This workload pod will contain:
  - Zero or more `initContainers`
  - The primary workload container
    - In many cases, this can be a reference directly to a published docker container
      - eg `golang:1.13` or `node:14.8.0-alpine3.11`
    - Some `BuildEnvironment`s may define several containers - for example, integration tests may need a database container as well.
      - See [BuildEnvironments](3-build_environments.md) for details on creating customized environments
  - One or more sidecar containers
    - These run alongside the primary workload container and mount a shared filesystem to make tools and functionality available to the job.
    - The `Buildkite Agent` sidecar container will always be injected into every pod.
      - This contains the `buildkite-agent` binary and the environment needed to use it.
    - There is a set of sidecars provided out of the box. Those, as well as custom sidecars, are documented [here](3-build_environments.md).

## BuildJob Configuration Options

TODO actual docs on options, examples, etc.

In the meantime, see the [BuildJob spec](../../api/v1alpha1/buildjob_types.go) for configuration options.

[Next Step: Build Environments](3-build_environment.md)
