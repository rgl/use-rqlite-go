#!/bin/bash
SCRIPT_PATH="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" &>/dev/null && pwd)"
source "$SCRIPT_PATH/env.sh"

# see https://github.com/rqlite/helm-charts/releases
# renovate: datasource=helm depName=rqlite registryUrl=https://rqlite.github.io/helm-charts
rqlite_chart_version='1.15.0'

echo "Creating $CLUSTER_NAME k8s..."
kind create cluster \
    --name="$CLUSTER_NAME" \
    --config="$SCRIPT_PATH/config.yml"
kubectl cluster-info

echo 'Creating the docker registry...'
# TODO create the registry inside the k8s cluster.
docker run \
    -d \
    --restart=unless-stopped \
    --name "$CLUSTER_NAME-registry" \
    --env REGISTRY_HTTP_ADDR=0.0.0.0:5001 \
    -p 5001:5001 \
    registry:2.8.3 \
    >/dev/null
while ! wget -q --spider http://localhost:5001/v2; do sleep 1; done;

echo 'Connecting the docker registry to the kind k8s network...'
# TODO isolate the network from other kind clusters with KIND_EXPERIMENTAL_DOCKER_NETWORK.
#      see https://github.com/kubernetes-sigs/kind/blob/v0.26.0/pkg/cluster/internal/providers/docker/network.go
docker network connect \
    kind \
    "$CLUSTER_NAME-registry"

echo 'Installing the rqlite helm chart...'
helm repo add rqlite https://rqlite.github.io/helm-charts
helm repo update
# search the chart and app versions, e.g.: in this case we are using:
#   NAME           CHART VERSION  APP VERSION  DESCRIPTION                                       
#   rqlite/rqlite  1.15.0         8.36.5       The lightweight, distributed relational databas...
helm search repo rqlite/rqlite --versions | head -3
# set the rqlite configuration.
cat >rqlite-values.yml <<EOF
replicaCount: 3
persistence:
  size: 1Gi
resources:
  requests:
    cpu: 500m
EOF
# install rqlite.
helm upgrade --install \
  rqlite \
  rqlite/rqlite \
  --version "$rqlite_chart_version" \
  --create-namespace \
  --namespace default \
  --values rqlite-values.yml \
  --wait
