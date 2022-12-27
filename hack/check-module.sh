#!/bin/bash

shaId=$(git log --before="1 week ago" --pretty=format:"%h" -n 1)
echo $shaId
files_changed=$(git diff $shaId --name-only "cmd/" "pkg/" "test/" | grep ".go" )

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
