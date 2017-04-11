#!/bin/bash
# Run the connector 
set -e

# trap terminate signal and pass to cocoon process
trap 'kill -TERM $CPID' SIGTERM SIGINT

# Set up go environment
export GOROOT=/go
export GOPATH=/gocode
PATH=$PATH:$GOROOT/bin:$GOPATH/bin
mkdir -p $GOPATH/bin

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
printf "> Building cocoon"
cd cocoon
rm -rf .glide/ && rm -rf vendor
glide --debug install
go build -v -o $GOPATH/bin/cocoon core/main.go

# start connector, store its process id and wait for it.
printf "Running Cocoon Connector"
cocoon connector & 
CPID=$!
wait $CPID
wait $CPID
EXIT_STATUS=$?