#!/usr/bin/env bash
set -euo pipefail

# 1) create kind cluster (if not exists)
kind get clusters | grep -q devops-challenge || kind create cluster --name devops-challenge --config kind-config.yaml

# 2) install calico for NetworkPolicy enforcement
kubectl apply -f https://raw.githubusercontent.com/projectcalico/calico/v3.26.0/manifests/calico.yaml

# 3) build images
./scripts/build-images.sh

# 4) install helm chart (install or upgrade)
helm upgrade --install devops-demo charts/devops-demo --wait --timeout 5m
echo "Deployment complete. Web service should be available on http://localhost:30080"
