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
TEST_TAGS=integration
COVERAGE_FN="coverage.out"

coverage_results="$PWD/coverageTmp"
test_results=./testing/json_results.log
mkdir -p $coverage_results
echo -n "" >$test_results

# build the payment-vault api as dependency for the integration tests which is required
mkdir -p ./payment-vault/api/publish/config
go build -o "./payment-vault/api/publish/" "./payment-vault/api/" 
cp ./payment-vault/api/config/* ./payment-vault/api/publish/config

cleanup() {
    echo "Cleaning up containers"
    docker compose -f ./project/docker-compose-infra-test.yml -p payment-test down
}

run_tests() {
  integration_test_dir=$1
  app_folder=$(dirname $integration_test_dir)
  app_name=$(basename $app_folder)
  echo "Start Integration Tests for $app_name"
  # Initialize the coverage file
  echo "mode: set" > $app_folder/$COVERAGE_FN
  for dir in $app_folder/*/; do
    if [ -d "$dir" ]; then
      echo "Start Test $dir"
      # -coverpkg should be parent folder to include coverage of packages from other modules
      go test -C=$dir -tags=$TEST_TAGS -timeout=$TEST_TIME_OUT -coverprofile=coverage.tmp -coverpkg=../... -json ./... >> $test_results
      if [ -f $dir/coverage.tmp ]; then
        tail -n +2 $dir/coverage.tmp >> $app_folder/$COVERAGE_FN
        rm $dir/coverage.tmp
      fi
    fi
  done
  # generate the html coverage report for each file
  go tool cover -html=$app_folder/coverage.out -o $coverage_results/$app_name.html
  # generate the coverage report for each funcion 
  go tool cover -func=$app_folder/coverage.out -o $coverage_results/$app_name.txt
  rm $app_folder/coverage.out
  echo "End Test $app_name"
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



# list all folders name integration-tests and ignore any permission error
test_folders=$(find . -type d -name "integration-tests" 2>/dev/null)
for f in $test_folders; do
    # set up tables and topics for the integration tests
    go run ./project/setup_infra.go
    set +e
    echo $f
    run_tests $f
    set -e
done

# format the test results into JUNIT
go run gotest.tools/gotestsum@latest --junitfile ./testing/results.xml --raw-command cat $test_results

# check any test failed
if grep -q '"Action":"fail"' $test_results; then
    echo "Test failed"
    exit 1
fi