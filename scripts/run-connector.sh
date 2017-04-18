#!/bin/bash
# Run the connector 
set -e

printf "> Starting Cocoon\n"

# Pull cocoon source
export GOPATH=/go
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
printf "> Running Cocoon Connector\n"
exec cocoon connector