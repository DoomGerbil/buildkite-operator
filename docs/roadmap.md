# Extremely Rough and Vague Roadmap

That will definitely not be adhered to.

## M1 (v0.1.0)

Minimally Usable Proof of Concept.

At this stage, this isn't really something you _should_ use, but it's enough that you _can_
run something so that we can begin to prove that it works.

- [ ] BuildJob Schema
- [ ] BuildJob Controller
- [ ] Buildkite API Client
- [ ] Core BuildEnvironment Schema
- [ ] buildkite-agent bootstrapper

## M2 (v0.2.0)

Configurable Environment

This stage starts to add in the features around running jobs in an environment that captures
the actual requirements to run actual tasks.

- [ ] BuildEnvironment schema extension
- [ ] Sidecar injection framework
- [ ] JobSpawner Controller
- [ ] Job Targeting Rules

## M3 (v0.3.0)

Usability and Productionization

This stage starts to add in real usability and production features. This is where we start to tape
over the sharp, rusty knife handles so that now we're using our sharp instruments with at least a duct
tape handle.

- [ ] Buildkite Helper Plugin
- [ ] First Sidecar container
- [ ] RepoMirror framework

## M4 (v0.4.0)

Flexibility and extensibility

This stage starts adding in the flexibility to allow users to inject their own capabilties.

- [ ] Custom Sidecar framework
- [ ] RepoSyncer Service

## M5 (v0.5.0)

Advanced Features

This stage is where we start adding in the functionality that woould be required for more diverse,
heterogenous, and complex CI environments.

- [ ] Validating/Mutating Webhooks
- [ ] Node Targeting/Affinity/Taints/Tolerances

## And more...
