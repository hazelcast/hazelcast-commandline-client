#! /bin/bash

# requires: jq

set -eu

start_queue=__datamigration_start_queue
status_map=__datamigration_1

clc queue -n $start_queue destroy --yes -q

echo "Waiting for the migration job to be available."
while true; do
	job=$(clc queue -n $start_queue poll -q)
	if [ "$job" != "-" ]; then
		break
	fi
	sleep 2
done

migration_id=$(echo $job | jq -r .migrationId)
status_map="__datamigration_${migration_id}"

echo "Received the migration job."
echo "Started the migration."
clc map -n $status_map set status -v json '{"status":"IN_PROGRESS"}' -q
sleep 10
clc map -n $status_map set 	status -v json '{"status":"COMPLETED"}' -q
echo "Migration ended."

