#!/bin/bash
# Run the connector 
set -e

printf "> Starting Cocoon\n"

# term_connector() {
#     if [ $cpid -ne 0 ]; then
#         kill -SIGTERM "$cpid"
#         wait "$cpid"
#     fi
#     exit 143;
# }

# trap terminate signal and pass to cocoon process
# trap 'kill ${!}; term_connector' SIGTERM SIGINT

# Set up go environment
export GOPATH=/go

# Pull cocoon source
branch=$VERSION
repoOwner=github.com/ncodes
repoOwnerDir=$GOPATH/src/$repoOwner
mkdir -p $repoOwnerDir
cd $repoOwnerDir
printf "> Fetching cocoon source. [branch=$branch] [dest=$repoOwnerDir]\n"
rm -rf cocoon
git clone --depth=1 -b $branch https://$repoOwner/cocoon

# build the binary
printf "> Building cocoon\n"
cd cocoon
rm -rf .glide/ && rm -rf vendor
glide --debug install
go build -v -o $GOPATH/bin/cocoon core/main.go

# start connector, store its process id and wait for it.
printf "Running Cocoon Connector\n"
exec cocoon connector
# cpid=$!

# while true
# do
#   tail -f /dev/null & wait ${!}
# done