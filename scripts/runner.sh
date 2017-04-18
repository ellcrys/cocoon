#!/usr/bin/env bash
# Runner.sh fetches the relevant run script for
# the connector. 
set -e
printf "> Runner.sh has started\n"

# Fetch the script and run it.
rm -f $RUN_SCRIPT_NAME
wget $RUN_SCRIPT_URL
chmod +x $RUN_SCRIPT_NAME
exec ./$RUN_SCRIPT_NAME