# Repo Mirroring

## Reasons for Keeping a Repo Mirror

In a cloud-provider Kubernetes environment, standard persistent disks usually cannot be mounted as `ReadWrite` by more than one `Pod` at
once- they can be used as `ReadOnceMany` - where one `Pod` has write access, and all others have read-only access, but not `ReadWriteMany`.

That means that there can only be one "owner" of the local mirror responsible for keeping the mirror updated, and all other users must not try to change the mirror.

In this system, that is the [`reposyncer` service](5-controllers_services.md#reposyncer), which mounts the repo mirror disk as read-write,
where all other pods using it mount the disk read-only and clone the remote repo as a shared reference to the mirror like so:

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

Because the mirror is mounted into each `Pod` as a persistent disk, the performance is far greater than doing a repo clone/fetch from the true remote.

## RepoSyncer

In order to speed up repo syncing, primary workload containers will mount the mirror disk read-only and reference the mirror when cloning.

However, those workers will still need to fetch the delta between the mirror and `HEAD@remote` over the internet, so the smaller this delta, the better the performance.

This component is responsible for keeping those persistent repo mirrors as up to date as possible. In the normal case, we could let `buildkite-agent` keep these mirrors up to date, but in this case, we want to have a single persistent repo mirror shared between all ephemeral containers - since `BuildJob` `Pod`s are destroyed after each job completes, they cannot have any persistent cross-job repo cache.
