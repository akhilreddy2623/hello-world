# lint configuration is in .golangci.yml
export GOPROXY="https://artifactory-pd-infra.aks.aze1.cloud.geico.net/artifactory/api/go/mvp-billing-golang-all,direct"
export GONOSUMDB="github.com/geico-private,dev.azure.com"
export CGO_ENABLED=0  

cli_version="v1.58.1"
# install golangci-lint if not existed in the environment yet
if [ -x "$(command -v golangci-lint)" ]; then
    installed_version=$(golangci-lint --version | awk '{print $4}')
    if [ "$installed_version" = "$cli_version" ]; then
        echo "golangci-lint $cli_version is already installed"
    else 
        echo "golangci-lint $installed_version is installed, but $cli_version is required"
        go install github.com/golangci/golangci-lint/cmd/golangci-lint@${cli_version}
    fi
else
    echo "golangci-lint is not installed"
    go install github.com/golangci/golangci-lint/cmd/golangci-lint@${cli_version}
fi


mods=$(go list -f '{{.Dir}}' -m)

for mod in $mods; do
    echo "Start checking $mod"
    golangci-lint run "$mod/..."
done