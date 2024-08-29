#!/bin/sh
BASEDIR=$(dirname "$0")
cd  "$BASEDIR"
echo "$BASEDIR"
export GOPROXY="https://artifactory-pd-infra.aks.aze1.cloud.geico.net/artifactory/api/go/mvp-billing-golang-all,direct"
export CGO_ENABLED=0
export GOOS=linux

go build -o "./../../../config-manager/api/publish/" "./../../../config-manager/api/." 

mkdir -p ./../../../config-manager/api/publish/config
cp ./../../../config-manager/api/config/* ./../../../config-manager/api/publish/config