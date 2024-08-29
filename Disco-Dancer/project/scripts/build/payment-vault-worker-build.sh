#!/bin/sh
BASEDIR=$(dirname "$0")
echo "$BASEDIR"
cd  "$BASEDIR"
export GOPROXY="https://artifactory-pd-infra.aks.aze1.cloud.geico.net/artifactory/api/go/mvp-billing-golang-all,direct"
export CGO_ENABLED=0

go build -o ./../../../payment-vault/worker/publish/ -ldflags="${LDFLAGS[*]}" "./../../../payment-vault/worker/." 

mkdir -p ./../../../payment-vault/worker/publish/config
cp ./../../../payment-vault/worker/config/* ./../../../payment-vault/worker/publish/config