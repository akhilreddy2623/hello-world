#!/bin/sh
BASEDIR=$(dirname "$0")
cd  "$BASEDIR"
echo "$BASEDIR"
export GOPROXY="https://artifactory-pd-infra.aks.aze1.cloud.geico.net/artifactory/api/go/mvp-billing-golang-all,direct"
export CGO_ENABLED=0
export GOOS=linux

go build -o "./../../../task-manager/api/publish/" "./../../../task-manager/api/." 

mkdir -p ./../../../task-manager/api/publish/config
cp ./../../../task-manager/api/config/* ./../../../task-manager/api/publish/config