#! /bin/bash

while true; do
  read -r line
  state=$(echo $line | cut -f1 -d' ')
  if [ "$state" == "REM" ]; then
    addr=$(echo $line | cut -f2 -d' ')
    uuid=$(echo $line | cut -f3 -d' ')
    echo "WARNING: Member $uuid at address $addr was removed from the cluster" | tee -a /tmp/warnings.txt
  fi
done