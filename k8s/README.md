# K8S

## Prerequisites

- kubectl (v1.31.2)
- kustomize (v5.4.2)
- k3d (v5.7.5)
- helm (v3.16.2)
- helmfile (v1.0.0-rc.8)

## Installation

- `k3d cluster create --config k3d-cluster.yaml`
- `helmfile -e dev init`
- `helmfile -e dev apply`
- `kubectl kustomize overlays/dev | kubectl apply -f -`

## Updating cluster

#### Helm changes

- `helmfile deps`
- `helmfile sync`

#### Kustomization changes

- `kubectl kustomize overlays/dev | kubectl apply -f -`
