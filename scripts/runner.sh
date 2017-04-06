#!/usr/bin/env bash
# Runner.sh fetches the relevant run script for
# the connector. 
set -e

echo "Runner.sh started"
trap 'echo Receive itttt' SIGTERM SIGINT
# Fetch the script and run it.
rm -f $RUN_SCRIPT_NAME
wget $RUN_SCRIPT_URL
while true; do :; done
chmod +x $RUN_SCRIPT_NAME
./$RUN_SCRIPT_NAME