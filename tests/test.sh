#! /bin/bash

set -e

expected_lines=1000

./map_put_all.sh
lines=$(./map_get_all.sh | python3 compare.py)

if [ "$lines" -ne "$expected_lines" ]; then
  echo "Expected: $expected_lines, got: $lines"
  exit 1
fi

echo "OK"