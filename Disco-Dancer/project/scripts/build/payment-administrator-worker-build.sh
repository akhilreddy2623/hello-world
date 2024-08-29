#!/bin/sh
BASEDIR=$(dirname "$0")
echo "$BASEDIR"
cd  "$BASEDIR"
export GOPROXY="https://artifactory-pd-infra.aks.aze1.cloud.geico.net/artifactory/api/go/mvp-billing-golang-all,direct"
export CGO_ENABLED=0

go build -o ./../../../payment-administrator/worker/publish/ -ldflags="${LDFLAGS[*]}" "./../../../payment-administrator/worker/." 

mkdir -p ./../../../payment-administrator/worker/publish/config
cp ./../../../payment-administrator/worker/config/* ./../../../payment-administrator/worker/publish/config