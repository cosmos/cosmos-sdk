# Starship

Starship is a tool to help simulate real chains, relayers and the interchain infra in a k8s cluster, locally, in the CI and
on large k8s clusters.

Perfect for running internal devnets, and writing e2e tests. Starship helps take projects from development to production.

## Directory Structure

* `configs/` is a directory holding various kinds of configuration for running Starship
  * `configs/local.yaml` is configured specially to run locally
  * `configs/devnet.yaml` is a larger scale dev environment that will spin up nodes in k8s cluster and keep them running for a week (after a weeks time nodes might start to fail)
  * `configs/ci.yaml` light way setup supposed to be run in the CI runner itself
* `scripts/`: handy scripts to interact with Starship live here. These are standalone scripts that can be run anywhere
* `tests/`: Directory holding the e2e tests to run against the system
* `Makefile`: Handy commands go here

## Getting Started

### Setup dependencies

Checkout the [docs](https://starship.cosmology.tech/get-started/step-1) or just run:
```bash
make setup-deps
```

### Connect to kubernetes cluster

Connect to a k8s cluster. Follow one of the 2 steps based on your operating system
* [2.1.1](https://starship.cosmology.tech/get-started/step-2#211-setup-with-kind-cluster): Spin up using kind cluster, for linux
* [2.1.2](https://starship.cosmology.tech/get-started/step-2#212-setup-with-docker-desktop): Using docker-desktop, for mac

Run following to check the connection
```bash
kubectl get nodes
```

### Startup

We use helm-charts for packaging all k8s based setup. This is controlled by the helm chart versions.
Run the following to fetch the helm chart from the `Makefile` variable `HELM_VERSION`.
```bash
make setup-helm
```

Now you can spin up the local cluster with:
```bash
make install
```

Check the pods are in running state with:
```bash
kubectl get pods
```

Once all the pods are in `Running` state run port-forward command to get all ports forwarded locally:
```bash
make port-forward
```

Enjoy!!!!!

### Teardown

Once you are done:
```bash
make stop
```

If you spun up the kubernetes cluster, then please stop it or if you used kind run `make clean-kind`

## Run tests

Tests are designed such that one can re-run the same tests against an already running infra.
This will save the cost of initialization of the infra.

Startup the cluster
```bash
make install

## check status of the pods
kubectl get pods

## Once the pods are up run:
make port-forward
 
## Run tests, can run this now multiple times as long as tests are running
make test

## Cleanup
make stop
```

## Troubleshooting local setup

Currently, there seems to be some issues when running starship on a local system. This section will help clear out some of them

### Not starting
If all or some of the pods are in `Pending` state, then it means that resources for docker containers are not enough.
There can be 2 ways around this:

1. Increase the resources for your local kubernetes cluster.
  * Docker Desktop: Go to `Settings` > `Resources`, increase CPU and memory
2. Reduce the resources for each of the nodes in `configs/local.yaml` file. You can look at `configs/ci.yaml` to understand the `resource` directive in the chains
  * `configs/ci.yaml` uses very little resources, so should be able to run locally

> NOTE: When resoureces are reduced or if the devnet has been running for a longer time, then the pods seem to die out or keep restarting. This is due to memory overflow. Will be fixed soon. For now
> one can just run
```bash
make delete

## wait for nodes to die out, check with
kubectl get pods

## restart
make install
```

### Long startup time


## Future Work

### Dev-UX
* We will get rid of the whole `scripts/` dir and replace it with a handy `starship` cli tool.
* Local infra spinup, specially on a Mac takes alot of time, this is something we will speedup much more
* We will have auto resource allocation, so manual troubleshooting is reduced
* Longer running pods or infra without issues.

### Testing/Infra
* Add more tests based on requirements, port all existing adhoc tests to Starship
* Add build system into starship (premetive exists) to be able to build the current simapp from the branch
* Start from non-empty genesis
* Add the concept of `jobs` into Starship config, where we can run predefined jobs against the infra
