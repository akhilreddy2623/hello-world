#!/bin/bash
set -e
echo working directory: $PWD
# run under root directory of the project
if [ $(basename $PWD) != "geico-payment-platform" ]; then
    echo "working directory should be root directory of the project"
    exit 1
fi

export GOPROXY="https://artifactory-pd-infra.aks.aze1.cloud.geico.net/artifactory/api/go/mvp-billing-golang-all,direct"
export GONOSUMDB="github.com/geico-private,dev.azure.com,artifactory-pd-infra.aks.aze1.cloud.geico.net/geico"
export CGO_ENABLED=0

TEST_TIME_OUT=180s
test_results=./testing/json_results.log
echo -n "" >$test_results

# build the payment-vault api as dependency for the integration tests
mkdir -p ./payment-vault/api/publish/config

go build -o "./payment-vault/api/publish/" "./payment-vault/api/" 

cp ./payment-vault/api/config/* ./payment-vault/api/publish/config


cleanup() {
    echo "Cleaning up containers"
    docker compose -f ./project/docker-compose-infra-test.yml -p payment-test down
}
trap cleanup EXIT
# tear down the previous running containers if any
cleanup

# start the containers needed for integration testing
docker compose -f ./project/docker-compose-infra-test.yml -p payment-test up -d

# wait for the containers to be ready
echo "Waiting 15s for the containers to be ready"
sleep 15

# clean the cache for the tests
go clean -testcache


# list all folders name integration-tests and ignore any permissions error
test_folders=$(find . -type d -name "integration-tests" 2>/dev/null)
for f in $test_folders; do
    # get the basename of parent folder
    modName=$(basename $(dirname $f))
    echo "Start Integration Tests for $modName"
    # recreate tables and topics
    go run ./project/setup_infra.go
    set +e
    go test -C $f -v --tags=integration -timeout $TEST_TIME_OUT -json ./... >>$test_results  
    echo "End Test $modName"
    set -e
done

# format the test results into JUNIT
go run gotest.tools/gotestsum@latest --format testname --junitfile ./testing/results.xml --raw-command cat $test_results

# check any test failed
if grep -q '"Action":"fail"' $test_results; then
    echo "Test failed"
    exit 1
fi