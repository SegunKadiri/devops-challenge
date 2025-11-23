#!/usr/bin/env bash
set -euo pipefail
helm uninstall devops-demo || true
kind delete cluster --name devops-challenge || true
echo "cluster and helm release removed"
