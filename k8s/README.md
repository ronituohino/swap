# Kubernetes

This is an development version only!

## Prerequisites

- kubectl (v1.31.2)
- kustomize (v5.4.2)
- k3d (v5.7.5)
- helm (v3.16.2)
- helmfile (v1.0.0-rc.8)
  - After install do: `helmfile init`

## Installation

- `k3d cluster create --config k3d-cluster.yaml`
- `helmfile -e dev apply`
- `chmod +x ./local-deploy.sh && ./local-deploy.sh`

## Running crawler

- `./local-deploy.sh -w` - uses args defined in [webcrawler job](../k8s/base/manifests/webcrawler/job.yaml)

## Updating cluster

#### Helm changes

- `helmfile deps`
- `helmfile sync`

#### Kustomization or source changes

- `./local-deploy.sh`
