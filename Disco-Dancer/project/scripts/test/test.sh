#!/bin/bash
set -e
echo working directory: $PWD
# run under root directory of the project
if [ $(basename $PWD) != "geico-payment-platform" ]; then
    echo "Please run the script from the root directory of the project"
    exit 1
fi

export GOPROXY="https://artifactory-pd-infra.aks.aze1.cloud.geico.net/artifactory/api/go/mvp-billing-golang-all,direct"
export GONOSUMDB="github.com/geico-private,dev.azure.com"
export CGO_ENABLED=0
export GOOS="linux"


# clean the cache for the tests
go clean -testcache

test_results="./testing/json_results.log"
> $test_results


# Get the directories of all the modules
mods=$(go list -f '{{.Dir}}' -m)
# Loop through each module and append the test results with json format
for mod in $mods; do
    set +e
    echo "Start Test $mod"
    go test -C "$mod" -timeout 120s -json ./... >> $test_results   
    echo "End Test $mod"
    set -e
done

# format the test results into JUNIT
go run gotest.tools/gotestsum@latest --junitfile ./testing/results.xml --raw-command cat $test_results

# check any test failed
if grep -q '"Action":"fail"' $test_results; then
    echo "Test failed"
    exit 1
fi    