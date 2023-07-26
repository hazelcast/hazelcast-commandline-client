#! /bin/bash

set -eu

start_queue=__datamigration_start_queue
status_map=__datamigration_1

clc queue -n $start_queue destroy --yes -q
clc map -n $status_map destroy --yes -q

echo "Waiting for the migration job to be available."
while true; do
	job=$(clc queue -n $start_queue poll -q)
	if [ "$job" != "-" ]; then
		break
	fi
	sleep 2
done

echo "Received the migration job."
echo "Started the migration."
clc map -n $status_map set status IN_PROGRESS -q
sleep 10
clc map -n $status_map set 	status COMPLETE -q
echo "Migration ended."

