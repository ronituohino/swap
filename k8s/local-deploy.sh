#!/bin/bash

set -e

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
echo "Building and pushing Indexer image..."
docker build --tag ronituohino/swap-search:indexer-$UUID ../indexer/
docker push ronituohino/swap-search:indexer-$UUID

echo ""
echo "Building and pushing Webcrawler image..."
docker build --tag ronituohino/swap-search:webcrawler-$UUID ../webcrawler/
docker push ronituohino/swap-search:webcrawler-$UUID

echo ""
echo "Updating kustomization.yaml with new tags..."
sed -i "s/newTag: api/newTag: api-$UUID/" base/kustomization.yaml
sed -i "s/newTag: indexer/newTag: indexer-$UUID/" base/kustomization.yaml
sed -i "s/newTag: webcrawler/newTag: webcrawler-$UUID/" base/kustomization.yaml

echo ""
echo "Applying Kubernetes configuration..."
if ! kubectl kustomize overlays/dev | kubectl apply -f -; then
  echo ""
  echo "Kubernetes configuration failed, reverting tags..."
  sed -i "s/newTag: api-$UUID/newTag: api/" base/kustomization.yaml
  sed -i "s/newTag: indexer-$UUID/newTag: indexer/" base/kustomization.yaml
  sed -i "s/newTag: webcrawler-$UUID/newTag: webcrawler/" base/kustomization.yaml
  exit 1
fi

echo ""
echo "Reverting kustomization.yaml tags..."
sed -i "s/newTag: api-$UUID/newTag: api/" base/kustomization.yaml
sed -i "s/newTag: indexer-$UUID/newTag: indexer/" base/kustomization.yaml
sed -i "s/newTag: webcrawler-$UUID/newTag: webcrawler/" base/kustomization.yaml

echo ""
echo "Deployment completed successfully!"
