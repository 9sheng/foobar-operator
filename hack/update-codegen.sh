#!/usr/bin/env bash
set -euo pipefail

cd $GOPATH/src/k8s.io/kubernetes/staging/src/k8s.io/code-generator

./generate-groups.sh all \
  github.com/9sheng/foobar-operator/pkg/client/action \
  github.com/9sheng/foobar-operator/pkg/apis \
  "action:v1"
