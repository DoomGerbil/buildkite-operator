# Controllers and Services

There are multiple controllers and services that run as part of the Operator, responsible for various bits and pieces of functionality.

Overall, the controller is largely responsible for ingesting the resource, creating the `Pod`, and then handing off
control to the `buildkite-agent` bootstrap in the container.

After that, the controller's responsibility is mostly to ensure that the container gets cleaned up after the Job completes or times out.

## JobSpawner

Having to run a separate agent outside of the Operator is
fine for a proof of concept, but overall adds complexity and more points of failure.

Enter the `JobSpawner` service - it effectively behaves as a buildkite-agent, and registers with the BuildKite service using an agent token provided in a `ConfigMap`.

When this `agent` is assigned a job by BuildKite, it will simply create `BuildJob` resources in the cluster based on the step data.

This serves a few purposes:

- Users no longer need to generate and create `BuildJob` K8S resources themselves.
  - So using a Buildkite Plugin is no longer necessary.
- Users running a job no longer need to have any access to the cluster

Thus, the end-user workflow changes to instead be:

1. A BuildKite pipeline runs, and creates one or more BuildKite jobs targeting the `JobSpawner` agent.
1. The `JobSpawner` creates `BuildJob` resources in the cluster based on the settings specified.

And the rest of the workflow proceeds as before.

### Targeting BuildKite Jobs at the JobSpawner

The `JobSpawner` registers as a BuildKite agent, and as such, standard [BuildKite agent targeting rules] apply.

By default, `JobSpawner` will register with a single label - `queue=bk-operator` - and can thus be targeted like so:

```yaml
steps:
  - command: "build.sh"
    agents:
      queue: bk-operator
```

#### JobSpawner Labels

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

#### Passing in config options

The only way to supply arbitrary parameters or options to a command step is via the agent labels.
Since we actually want to route all  so we don't have a straightforward way to provide options to
the operator directly.

TBD what the right answer here is, though I lean towards Option 2.

##### Option 1 - Environment variables

Any further options to be specified must be written into the environment variables for the step it should apply to.

For example:

```yaml
steps:
  - command: "build.sh"
    agents:
      queue: bk-operator
    env:
      BK_OPERATOR_BUILD_ENVIRONMENT: name-of-build-environment
      BK_OPERATOR_LOG_PATHS: /path/to/log/file/1.log;/path/to/log/file/2.log
```

##### Option 2 - Buildkite Plugin Helper

Options for the build will be provided via a `bk-operator` BuildKite plugin, like so:

```yaml
steps:
  - command: "build.sh"
    agents:
      queue: bk-operator
    plugins:
    - buildkite-operator#v0.0.1:
        build_environment: name-of-build-environment
        log_paths:
        - "/path/to/log/file/1.log"
        - "/path/to/log/file/2.log"
```

The plugin's implementation is TBD, but it will basically use the more user-friendly arguments to
construct the `BuildJob` resource and insert it into the cluster.

### Implementation details

Most details are TBD, but this could take a few shapes.

- Maybe the `JobRunner` registers directly as an agent that speaks the buildkite-agent protocol?
- Maybe it runs a real `buildkite-agent` but overrides the all of the hooks so that the only thing
jobs run on it can do is to translate and upload `BuildJob` resources.

## Validating and Mutating Webhooks

The `BuildJobValidatingWebhook` will reject invalid job resources at creation time, ensuring that jobs
with impossible-to-schedule specs will be rejected with a meaningful error message.

This should be considered best effort, since there will always be some unvalidatable things.

TODO (future): Provide some hook for allowing custom, user-defined validation.

[BuildKite agent targeting rules]: https://buildkite.com/docs/agent/v3/cli-start#agent-targeting
