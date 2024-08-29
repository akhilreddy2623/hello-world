#!/bin/sh
BASEDIR=$(dirname "$0")
echo "$BASEDIR"
cd  "$BASEDIR"
export GOPROXY="https://artifactory-pd-infra.aks.aze1.cloud.geico.net/artifactory/api/go/mvp-billing-golang-all,direct"
export CGO_ENABLED=0

go build -o ./../../../payment-executor/worker/publish/ -ldflags="${LDFLAGS[*]}" "./../../../payment-executor/worker/." 

mkdir -p ./../../../payment-executor/worker/publish/config
cp ./../../../payment-executor/worker/config/* ./../../../payment-executor/worker/publish/config