#!/bin/bash
# Run the connector 
set -e


# Set up go environment
export GOROOT=/go
export GOPATH=/gocode
PATH=$PATH:$GOROOT/bin:$GOPATH/bin
mkdir -p $GOPATH/bin

# Pull cocoon source
branch="master"
repoOwner=github.com/ncodes
repoOwnerDir=$GOPATH/src/$repoOwner
mkdir -p $repoOwnerDir
cd $repoOwnerDir
printf "> Fetching cocoon source. [branch=$branch] [dest=$repoOwnerDir]\n"
git clone --depth=1 -b $branch https://$repoOwner/cocoon

# build the binary
printf "> Building cocoon"
cd cocoon
rm -rf .glide/ && rm -rf vendor
glide --debug install
/go/bin/go build -v -o $GOPATH/bin/cocoon core/main.go

# start connector 
printf "Running Cocoon Connector"
cocoon connector