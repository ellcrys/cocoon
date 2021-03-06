#!/bin/bash
# Run the connector 
set -e
export GOPATH=/go
printf "> Starting Cocoon [env=$ENV]\n"

# Fetch and build connector from source in dev/test environemtn
if [ $ENV != "production" ]; then
    branch=$VERSION
    repoOwner=github.com/ellcrys
    repoOwnerDir=$GOPATH/src/$repoOwner
    mkdir -p $repoOwnerDir
    cd $repoOwnerDir
    printf "> Fetching cocoon source. [branch=$branch] [dest=$repoOwnerDir]\n"
    rm -rf cocoon
    git config --global core.compression 0
    git clone --depth=1 -b $branch https://$repoOwner/cocoon

    # build the binary
    printf "> Building connector\n"
    cd cocoon
    govendor fetch -v +out
    go build -v -o $GOPATH/bin/connector core/connector/main.go
else
    # Fetch pre-built binary 
    fileName="connector_${VERSION}.zip"
    rm -rf $fileName
    rm -rf $GOPATH/bin/connector
    printf "> Downloading pre-built binary [version: $VERSION, date: $(date)]\n"
    wget "https://storage.googleapis.com/hothot/${fileName}"
    unzip $fileName
    mv connector $GOPATH/bin/connector
    printf "Checksum: $(shasum $GOPATH/bin/connector)\n"
    
    # Give the cocoon code container some time to start
    sleep 5
fi

# start connector
printf "> Running Connector\n"
exec connector start