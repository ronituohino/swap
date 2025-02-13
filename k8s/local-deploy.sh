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
docker build --tag ronituohino/swap-search:api-$UUID ../api/
docker push ronituohino/swap-search:api-$UUID

echo ""
echo "Building and pushing IDF image..."
docker build --tag ronituohino/swap-search:idf-$UUID ../idf/
docker push ronituohino/swap-search:idf-$UUID

echo ""
echo "Building and pushing Indexer image..."
docker build --tag ronituohino/swap-search:indexer-$UUID ../indexer/
docker push ronituohino/swap-search:indexer-$UUID

if [ "$WEBCRAWLER_ENABLED" = true ]; then
  echo ""
  echo "Building and pushing Webcrawler image..."
  docker build --tag ronituohino/swap-search:webcrawler-$UUID ../webcrawler/
  docker push ronituohino/swap-search:webcrawler-$UUID
else
  echo ""
  echo "Skipping Webcrawler build and push..."
fi

echo ""
echo "Updating kustomization.yaml with new tags..."
sed -i "s/newTag: api/newTag: api-$UUID/" base/kustomization.yaml
sed -i "s/newTag: idf/newTag: idf-$UUID/" base/kustomization.yaml
sed -i "s/newTag: indexer/newTag: indexer-$UUID/" base/kustomization.yaml
sed -i "s/newTag: webcrawler/newTag: webcrawler-$UUID/" base/kustomization.yaml

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
  sed -i "s/newTag: api-$UUID/newTag: api/" base/kustomization.yaml
  sed -i "s/newTag: idf-$UUID/newTag: idf/" base/kustomization.yaml
  sed -i "s/newTag: indexer-$UUID/newTag: indexer/" base/kustomization.yaml
  sed -i "s/newTag: webcrawler-$UUID/newTag: webcrawler/" base/kustomization.yaml
  [ "$WEBCRAWLER_ENABLED" = false ] && sed -i 's/suspend: true/suspend: false/' base/manifests/webcrawler/job.yaml
  exit 1
fi

echo ""
echo "Reverting kustomization.yaml tags..."
sed -i "s/newTag: api-$UUID/newTag: api/" base/kustomization.yaml
sed -i "s/newTag: idf-$UUID/newTag: idf/" base/kustomization.yaml
sed -i "s/newTag: indexer-$UUID/newTag: indexer/" base/kustomization.yaml
sed -i "s/newTag: webcrawler-$UUID/newTag: webcrawler/" base/kustomization.yaml
[ "$WEBCRAWLER_ENABLED" = false ] && sed -i 's/suspend: true/suspend: false/' base/manifests/webcrawler/job.yaml

echo ""
echo "Deployment completed successfully!"
