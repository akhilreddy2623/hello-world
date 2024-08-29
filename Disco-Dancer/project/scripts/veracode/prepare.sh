#!/bin/bash
set -e
export GOPROXY="https://artifactory-pd-infra.aks.aze1.cloud.geico.net/artifactory/api/go/mvp-billing-golang-all,direct"
export GONOSUMDB="github.com/geico-private,dev.azure.com,artifactory-pd-infra.aks.aze1.cloud.geico.net/geico"
export CGO_ENABLED=0
# repo is structured as using go work space, so we need to append the local modules dependency to the go.mod file
# to make each module to be independentenly buildable with vendor mode required by veracode
mods=$(go list -f '{{.Path}}' -m)
requireLocalModules=""
replaceLocalModules=""
for modName in $mods; do
  path=$(go list -f '{{.Dir}}' -m $modName)
  requireStmt="require $modName v1.0.0"
  replaceStmt="replace $modName v1.0.0 => $path"

  requireLocalModules="$requireLocalModules\n$requireStmt"
  replaceLocalModules="$replaceLocalModules\n$replaceStmt"
done

appendContent="$requireLocalModules\n$replaceLocalModules"
echo -e "append content: $appendContent"

# 1. append the require and replace statements to the needed go.mod file
# 2. generate the vendor folder 
# 3. zip the scanning modules folder
scanningMods=('geico.visualstudio.com/Billing/plutus/payment-administrator-worker'
              'geico.visualstudio.com/Billing/plutus/payment-administrator-api'
              'geico.visualstudio.com/Billing/plutus/payment-executor-worker'
              'geico.visualstudio.com/Billing/plutus/payment-executor-api'
              'geico.visualstudio.com/Billing/plutus/task-manager-worker'
              'geico.visualstudio.com/Billing/plutus/task-manager-api'
              'geico.visualstudio.com/Billing/plutus/payment-vault-worker'
              'geico.visualstudio.com/Billing/plutus/payment-vault-api'
              'geico.visualstudio.com/Billing/plutus/config-manager-api'
              'geico.visualstudio.com/Billing/plutus/data-hub-worker'
             )
for modName in "${scanningMods[@]}"; do
  modFile=$(go list -m -f '{{.GoMod}}' $modName)
  echo -e "$appendContent" >> "$modFile"

  modPath=$(go list -m -f '{{.Dir}}' $modName)
  echo "generating vendor folder for $modName"
  go mod -C $modPath tidy
  go mod -C $modPath vendor
   # make sure the application can build successfully
  go build -C $modPath
  # zip is not available in the github runner for now, use go program to siumulate the zip command
  modBase=$(basename $modName)
  go run ./project/scripts/veracode/zip.go -src "$modPath" -o "./testing/$modBase.zip"
done

