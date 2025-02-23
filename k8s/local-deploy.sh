#!/bin/bash

set -e

WEBCRAWLER_ENABLED=false
while getopts "w" opt; do
  case $opt in
    w)
      WEBCRAWLER_ENABLED=true
      ;;
    \?)
      echo "Invalid option: -$OPTARG" >&2
      exit 1
      ;;
  esac
done

echo ""
echo "Starting deployment process..."

UUID=$(uuidgen)
echo ""
echo "Generated UUID: $UUID"

echo ""
echo "Building and pushing API image..."
docker build --tag swap/api:$UUID ../api/
k3d image import swap/api:$UUID -c swap-cluster

echo ""
echo "Building and pushing IDF image..."
docker build --tag swap/idf:$UUID ../idf/
k3d image import swap/idf:$UUID -c swap-cluster

echo ""
echo "Building and pushing Indexer image..."
docker build --tag swap/indexer:$UUID ../indexer/
k3d image import swap/indexer:$UUID -c swap-cluster

if [ "$WEBCRAWLER_ENABLED" = true ]; then
  echo ""
  echo "Building and pushing Webcrawler image..."
  docker build --tag swap/webcrawler:$UUID ../webcrawler/
  k3d image import swap/webcrawler:$UUID -c swap-cluster
else
  echo ""
  echo "Skipping Webcrawler build and push..."
fi

echo ""
echo "Updating kustomization.yaml with new tags..."
sed -i "s/newTag: uuid/newTag: $UUID/" base/kustomization.yaml

if [ "$WEBCRAWLER_ENABLED" = false ]; then
  echo ""
  echo "Setting webcrawler suspension to true..."
  sed -i 's/suspend: false/suspend: true/' base/manifests/webcrawler/job.yaml
fi

echo ""
echo "Applying Kubernetes configuration..."
kubectl delete job webcrawler -n swap-dev || true
if ! kubectl kustomize overlays/dev | kubectl apply -f -; then
  echo ""
  echo "Kubernetes configuration failed, reverting tags..."
  sed -i "s/newTag: $UUID/newTag: uuid/" base/kustomization.yaml
  [ "$WEBCRAWLER_ENABLED" = false ] && sed -i 's/suspend: true/suspend: false/' base/manifests/webcrawler/job.yaml
  exit 1
fi

echo ""
echo "Reverting kustomization.yaml tags..."
sed -i "s/newTag: $UUID/newTag: uuid/" base/kustomization.yaml
[ "$WEBCRAWLER_ENABLED" = false ] && sed -i 's/suspend: true/suspend: false/' base/manifests/webcrawler/job.yaml

echo ""
echo "Deployment completed successfully using version $UUID!"
