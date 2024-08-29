#!/bin/sh
BASEDIR=$(dirname "$0")
cd  "$BASEDIR"
echo "$BASEDIR"
export GOPROXY="https://artifactory-pd-infra.aks.aze1.cloud.geico.net/artifactory/api/go/mvp-billing-golang-all,direct"
export CGO_ENABLED=0
export GOOS=linux

go build -o "./../../../payment-administrator/api/publish/" "./../../../payment-administrator/api/." 

mkdir -p ./../../../payment-administrator/api/publish/config
cp ./../../../payment-administrator/api/config/* ./../../../payment-administrator/api/publish/config