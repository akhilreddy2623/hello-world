#!/bin/sh
BASEDIR=$(dirname "$0")
cd  "$BASEDIR"
echo "$BASEDIR"
export GOPROXY="https://artifactory-pd-infra.aks.aze1.cloud.geico.net/artifactory/api/go/mvp-billing-golang-all,direct"
export CGO_ENABLED=0
export GOOS=linux

go build -o "./../../../payment-executor/api/publish/" "./../../../payment-executor/api/." 

mkdir -p ./../../../payment-executor/api/publish/config
cp ./../../../payment-executor/api/config/* ./../../../payment-executor/api/publish/config