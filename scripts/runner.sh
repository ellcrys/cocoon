#!/usr/bin/env bash
# Runner.sh fetches the relevant run script for
# the connector. 
set -e
echo "Runner.sh has started"
trap 'echo caught the signal' SIGTERM SIGINT

# Fetch the script and run it.
rm -f $RUN_SCRIPT_NAME
wget $RUN_SCRIPT_URL
chmod +x $RUN_SCRIPT_NAME
./$RUN_SCRIPT_NAME &
PID=$!
wait $PID
wait $PID
EXIT_STATUS=$?