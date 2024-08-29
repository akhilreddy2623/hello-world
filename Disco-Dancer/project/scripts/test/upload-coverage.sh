#!/bin/bash
# DO NOT RUN IN LOCAL ENVIRONMENT
set -e
echo working directory: $PWD
# run under root directory of the project
if [ $(basename $PWD) != "geico-payment-platform" ]; then
    echo "working directory should be root directory of the project"
    exit 1
fi

BRANCH_NAME=gh-pages
svgTmpPath="./project/scripts/test/coverage.svg.tmp"
# find all text files
dataFiles=$(find $PWD/coverageTmp -type f -name "*.txt")
for f in $dataFiles; do
  set +e

  modName=$(basename $f .txt)
  echo "Starting to get coverage reports for $modName"
  # Extract total coverage: the decimal number from the last line of the function report.
  modCoverage=$(cat $f | tail -1 | grep -Eo '[0-9]+\.[0-9]')
  if [[ $modCoverage =~ ^[0-9]+([.][0-9]+)?$ ]]; then
    echo "Coverage for $modName is $modCoverage%"
    # generate the badge with the coverage rate
    cat $svgTmpPath | sed "s/\${coverageRate}/$modCoverage/g"  > $PWD/coverageTmp/$modName.svg
  else
    echo "No coverage for $modName"
  fi

  set -e
done
   
git config --global user.name "github-actions[bot]"
git config --global user.email "41898282+github-actions[bot]@users.noreply.github.com"
git fetch
git checkout $BRANCH_NAME
cd $PWD
rm -rf ./coverages
mv ./coverageTmp ./coverages
git add ./coverages
git commit --amend -m "Automated report"
git push -f