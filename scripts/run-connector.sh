#!/bin/bash
# Run the connector 
set -e
export GOPATH=/go

printf "> Starting Cocoon\n"

if [ $ENV != "production" ]; then
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
    printf "> Building connector\n"
    cd cocoon
    rm -rf .glide/ && rm -rf vendor
    glide install
    go build -v -o $GOPATH/bin/connector core/connector/main.go
else
    printf "> Downloading pre-built binary [version: $VERSION]"
    wget "https://storage.googleapis.com/krogan/connector_${VERSION}.zip"
    unzip "connector_${VERSION}.zip"
    mv connector $GOPATH/bin/connector
fi

# start connector, store its process id and wait for it.
printf "> Running Connector\n"
exec connector start