#!/bin/sh
BASEDIR=$(dirname "$0")
echo "$BASEDIR"
cd  "$BASEDIR"
export GOPROXY="https://artifactory-pd-infra.aks.aze1.cloud.geico.net/artifactory/api/go/mvp-billing-golang-all,direct"
export CGO_ENABLED=0

go build -o ./../../../data-hub/worker/publish/ -ldflags="${LDFLAGS[*]}" "./../../../data-hub/worker/." 

mkdir -p ./../../../data-hub/worker/publish/config
cp ./../../../data-hub/worker/config/* ./../../../data-hub/worker/publish/config