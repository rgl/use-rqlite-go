#!/bin/bash
set -euo pipefail

CLUSTER_NAME='use-rqlite-go'
export KUBECONFIG="$PWD/kubeconfig.yml"
