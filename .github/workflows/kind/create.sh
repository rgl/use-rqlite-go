#!/bin/bash
SCRIPT_PATH="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" &>/dev/null && pwd)"
source "$SCRIPT_PATH/env.sh"

# see https://github.com/rqlite/helm-charts/releases
# see https://artifacthub.io/packages/helm/rqlite/rqlite
# renovate: datasource=helm depName=rqlite registryUrl=https://rqlite.github.io/helm-charts
rqlite_chart_version='2.0.0'

# see https://hub.docker.com/_/registry
# see https://github.com/distribution/distribution/releases
# renovate: datasource=docker depName=registry
registry_image_version='3.0.0'

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
    --volume "$SCRIPT_PATH/registry-config.yml:/etc/distribution/config.yml:ro" \
    --env OTEL_SDK_DISABLED=true \
    --env OTEL_TRACES_EXPORTER=none \
    --env OTEL_METRICS_EXPORTER=none \
    --env OTEL_LOGS_EXPORTER=none \
    -p 5001:5001 \
    "registry:$registry_image_version" \
    >/dev/null
while ! wget -q --spider http://localhost:5001/v2/; do sleep 1; done;

echo 'Connecting the docker registry to the kind k8s network...'
# TODO isolate the network from other kind clusters with KIND_EXPERIMENTAL_DOCKER_NETWORK.
#      see https://github.com/kubernetes-sigs/kind/blob/v0.30.0/pkg/cluster/internal/providers/docker/network.go
docker network connect \
    kind \
    "$CLUSTER_NAME-registry"

echo 'Installing the rqlite helm chart...'
helm repo add rqlite https://rqlite.github.io/helm-charts
helm repo update
# search the chart and app versions, e.g.: in this case we are using:
#   NAME           CHART VERSION  APP VERSION  DESCRIPTION                                       
#   rqlite/rqlite  2.0.0          9.1.3        The lightweight, distributed relational databas...
helm search repo rqlite/rqlite --versions | head -3
# set the rqlite configuration.
cat >rqlite-values.yml <<EOF
image:
  tag: 9.1.3
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
