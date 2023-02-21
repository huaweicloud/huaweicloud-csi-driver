#!/bin/bash

vendor_changed=$(git diff ce821bcf5253 5e7aaafe --name-only "vendor/" | grep ".go" )
if [[ -n "$vendor_changed" ]]; then
  echo "Vendor files has changed, run all UT"
  echo "labels=EVS || SFS || SFS_TURBO || OBS" >> $GITHUB_OUTPUT
  echo "skip_e2e=false" >> $GITHUB_OUTPUT
  more $GITHUB_OUTPUT
  exit 0
fi

files_changed=$(git diff $1 $2 --name-only "cmd/" "pkg/" "test/" | grep ".go" )

echo "File(s) Changed:"
echo "$files_changed"

labels=""
function addTags(){
  if [[ -z "$labels" ]]; then
    labels=${1}
  else
    labels="${labels} || ${1}"
  fi
}

evs_tag=$(echo $files_changed | grep "evs" | wc -l)
if [[ "$evs_tag" -eq "1" ]]; then
  echo "EVS has changed, required to run the EVS E2E tests"
  addTags "EVS"
fi

sfs_tag=$(echo $files_changed | grep -E "sfs_|/sfs/" | wc -l)
if [[ "$sfs_tag" -eq 1 ]]; then
  echo "SFS has changed, required to run the SFS E2E tests"
  addTags "SFS"
fi

sfsturbo_tag=$(echo $files_changed | grep "sfsturbo" | wc -l)
if [[ "$sfsturbo_tag" -eq 1 ]]; then
  echo "SFS Turbo has changed, required to run the SFS Turbo E2E tests"
  addTags "SFS_TURBO"
fi

obs_tag=$(echo $files_changed | grep "obs" | wc -l)
if [[ "$obs_tag" -eq 1 ]]; then
  echo "OBS has changed, required to run the OBS E2E tests"
  addTags "OBS"
fi

if [[ -z "$labels" ]]; then
  labels="NONE"
  echo "skip_e2e=true" >> $GITHUB_OUTPUT
else
  echo "skip_e2e=false" >> $GITHUB_OUTPUT
fi
echo "labels=${labels}" >> $GITHUB_OUTPUT

more $GITHUB_OUTPUT
echo
echo "Module(s) changed: ${labels}"
echo

exit 0
