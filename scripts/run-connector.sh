#!/bin/bash
# Run the connector 
set -e
export GOPATH=/go
printf "> Starting Cocoon [env=$ENV]\n"

# Fetch and build connector from source in dev/test environemtn
if [ $ENV != "production" ]; then
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
    govendor fetch -v +out
    go build -v -o $GOPATH/bin/connector core/connector/main.go
else
    # Fetch pre-built binary 
    printf "> Downloading pre-built binary [version: $VERSION]\n"
    fileName="connector_${VERSION}.zip"
    rm -rf $fileName
    wget "https://storage.googleapis.com/krogan/${fileName}"
    unzip $fileName
    mv connector $GOPATH/bin/connector
    
    # Give the cocoon container some time to start
    sleep 5
fi

# start connector
printf "> Running Connector\n"
exec connector start