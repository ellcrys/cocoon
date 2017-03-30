# Runner.sh fetches the relevant run script for
# the connector. 
set -e

# Fetch the script and run it.
wget $RUN_SCRIPT_URL
bash $RUN_SCRIPT_NAME