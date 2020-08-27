# BuildKite Operator User's Guide

The core usage loop of the Operator is:

1. A `BuildJob` resource is created in the cluster.
1. The `buildjob-builder` controller service creates a pod based on the supplied spec.
    - The environment that this job runs is defined by a [`BuildEnvironment`](#BuildEnvironment).
1. The job runs inside this container, eventually exiting and reporting a status code.
    - This runs the standard buildkite-agent workflow, starting with `buildkite-agent bootstrap`.
    - Various init and sidecar containers are attached to the pod to provide necessary services.
1. When the primary workload container has completed, the controller will reap the completed resource.

## Requirements

Since users are responsible for creating `BuildJob` resources in the cluster
for every BuildKite job, there are two implicit dependencies:

1. Something or someone must be able to create the `BuildJob` resource in the cluster.
    - This could be done in several ways:
      - A BuildKite plugin that executes on a static BuildKite agent running outside of the context of the Operator
      - A command-line tool run directly by the user
      - A running service that accepts requests and creates the resources.
      - A Buildkite agent owned and run by the Operator solely responsible for creating `BuildJob` resources.
    - In all cases, the creating mechanism must have some level of access into the cluster to create the resources.
1. The BuildKite step definition has to be converted into a `BuildJob` resource for the Operator to schedule it.
    - Thus, every job must either:
      - Have a pre-created static `BuildJob` definition stored somewhere
      - Generate a dynamic `BuildJob` definition at runtime
      - Use a BuildKite Plugin to generate a `BuildJob` definition
    - And then the non-Operator agent would create the resource in the cluster.

Our proposal is the [JobSpawner service](5-controllers_services.md#JobSpawner) below for details on the proposed solution to this issue.

Until this implementation exists, `BuildJobs` can be manually inserted into the cluster via `kubectl apply -f` or the like to exercise the rest of the Operator's functionality.
