#! /bin/bash

set -e
set -u

path="${1:-/tmp/warnings.txt}"

while true; do
  read -r line
  state=$(echo $line | cut -f3 -d' ')
  echo "STATE: $state"
  if [ "$state" == "CLIENT_DISCONNECTED" ]; then
    echo "WARNING: Client disconnected" | tee -a "$path"
  elif [ "$state" == "REMOVED" ]; then
    addr=$(echo $line | cut -f4 -d' ')
    uuid=$(echo $line | cut -f6 -d' ')
    echo "WARNING: Member $uuid at address $addr was removed from the cluster" | tee -a "$path"
  fi
done