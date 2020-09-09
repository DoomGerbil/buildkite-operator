# BuildEnvironments

TODO: rethink this name. Perhaps a `BuildEnvironment` is composed or one or more `JobContainer` plus a configuration, and that plus the runtime injected additions compose the full `BuildEnvironment`.

This resource defines a container environment. This may be the environment in which a job will run, or
it may be a supplimental container that's used by other `BuildEnvironments` as a sidecar or an initContainer.

## Init Containers

One or more containers that run before the main workload, to ensure that pre-work is done before the job starts. These can do required setup, like syncing code, populating a test database, etc.

Some important notes about these:

- All `initContainers` are guaranteed to complete before the job starts.
- `initContainers` have write access to the same filesystem as the main workload container, _before_ the workload executes.
- All `initContainers` run in serial, and block workload startup, so adding several of these can have a negative performance impact.
- All `initContainer` failures are fatal, thus any failed `initContainer` job is inherently fatal to the `BuildJob`.

### RepoMirror initContainer

This container will check the cluster-local repo cache to see if the requested git repo is present, and will copy it into the pod if so.
It will also ensure that the pod's clone of the repo is up to date before the primary workload begins.

See the [RepoSyncer doc](4-repo_mirroring.md) for full details.

### ArtifactDownloader initContainer

This container will mount a volume in the `Pod` (default path `/artifacts-imported`), and download all specified pipeline artifacts into it.

## Sidecar Containers

Zero or more `sidecar` containers may be run in the `BuildJob` `Pod`, alongside the primary workload.

Key differences between `initContainers` and `sidecars`:

- `sidecar` containers are started at the same time as the primary workload.
- `sidecar` containers run in parallel and do not block execution
- `sidecar` containers _do_ block the `Pod` from _completing_ if they are still running.
- `sidecar` containers can fail without affecting the primary workload's exit status.
- `sidecar` containers exit status _will_ affect the `Pod`'s exit status.
- `sidecar` containers have write access to the `Pod` filesystem at the same time as the main workload, so `sidecar` containers must take care not to corrupt or interfere with the `BuildJob`'s working directory.

Only the `BuildKite Agent` sidecar is required. All other provided `sidecar` containers can be used or not, when needed.

Custom sidecars can also be defined and used as well.

### BuildKite-Agent Sidecar

This contains the `buildkite-agent` binary at a specified version, which will be mounted into the pod.

TODO/TBD: where? `/buildkite-agent`? On the PATH somewhere? `/usr/bin`?

It is also responsible for ensuring that the pod is authenticated to buildkite.com so that it can be
used for artifacts, annotations, meta-data, etc.

TODO: How does this inject the required environment variables into the main workload? Do we also need a `buildkite-auth` `initContainer` or something?

TODO: Does this also run the `buildkite-agent bootstrap`? If not, where do we do that, since we want to hook into that somewhere?

#### LogExporter Sidecar

This container can be attached to a job along with a per-job configuration, which chooses a named log sink and sets log source configuration.

##### Log Sinks

Available log sinks (eg `ElasticSearch`, `fluentd`, `stackdriver`, etc) are specified in a cluster-level `ConfigMap` that defines the necessary details for how to access them (endpoints, authentication credentials, etc), and sets a default log sink.

##### Log Sources

First, the sidecar will mount a volume in the `Pod` (default path `/logs`), and any files written here will be captured and exported.

Next, the `BuildJob` config can specify to capture stdout/stderr.

Finally, the `BuildJob` configuration can specify a list of filepaths, and the sidecar will watch for content to be written to those paths and export them.

#### ArtifactUploader Sidecar

This container will mount a volume in the `Pod` (default path `/artifacts`), and any files written under this path will be captured as a Buildkite artifact with the stored path relative to `/artifacts`.

#### MetricsExporter Sidecar

This container will mount a volume in the `Pod` (default root path `/metrics`) and export any metrics written to files in certain directories.

Export may be configured to happen via one or more of the following:

1. `prometheus-push`
    - Requires a pre-configured Prometheus `pushgateway`
    - Exports from `$METRICS_ROOT/prometheus-push`
    - The pushgateway config must be defined in a specified `ConfigMap`
    - All files written to this path must be valid Prometheus metrics
1. `prometheus` - Uses `node_exporter` to expose metrics on a specified port.
    - Exports from `$METRICS_ROOT/prometheus`
    - Prometheus must be configured to monitor the service (out of scope of this doc)
    - All files written to this path must be valid Prometheus metrics
1. `datadog`
    - TBD, but this is something that I suspect will be a common need.

#### Custom Sidecar Containers

Custom sidecar containers may be defined the same way as a `BuildEnvironment`.

For example, maybe you want to have a test result exporter that watches for the main workload to complete, and then collects a JUnit test report XML output and sends it to an external service for processing/reporting.

By putting this in a sidecar instead of in your primary workload container, it becomes both reusable across multiple jobs, and allows you to make changes to your test reporting without needing to change your build environment.

[Next Step: Repo Mirroring](4-repo_mirroring.md)
