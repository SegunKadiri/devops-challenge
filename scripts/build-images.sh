#!/usr/bin/env bash
set -euo pipefail

# Build the local monitor image referenced in values.yaml
docker build -t devops-monitor:local ./go-monitor
echo "Built devops-monitor:local"
