#!/bin/sh
BASEDIR=$(dirname "$0")
cd  "$BASEDIR"
echo "$BASEDIR"
#export GOPROXY="https://artifactory-pd-infra.aks.aze1.cloud.geico.net/artifactory/api/go/core-golang-all"
export GOPROXY="https://artifactory-pd-infra.aks.aze1.cloud.geico.net/artifactory/api/go/mvp-billing-golang-all,direct"
export GONOSUMDB="github.com/geico-private,dev.azure.com"
export GOPRIVATE=""
export CGO_ENABLED=0
export GOOS=linux


go build -o "./../../../payment-vault/api/publish/" "./../../../payment-vault/api/." 

mkdir -p ./../../../payment-vault/api/publish/config
cp ./../../../payment-vault/api/config/* ./../../../payment-vault/api/publish/config