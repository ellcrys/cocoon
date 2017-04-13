#!/usr/bin/env bash
# Runner.sh fetches the relevant run script for
# the connector. 
set -e
echo "Runner.sh has started"

term_handler() {
    if [ $pid -ne 0 ]; then
        kill -SIGTERM "$pid"
        wait "$pid"
    fi
    exit 143; 
}

trap 'kill ${!}; term_handler' SIGTERM SIGINT

# Fetch the script and run it.
rm -f $RUN_SCRIPT_NAME
wget $RUN_SCRIPT_URL
chmod +x $RUN_SCRIPT_NAME
./$RUN_SCRIPT_NAME &
pid=$!

while true
do
  tail -f /dev/null & wait ${!}
done