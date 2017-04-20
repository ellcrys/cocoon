# Run an Orderer. 
set -e 

VERSION=$ORDERER_VERSION

# Fetch and build Orderer from source in dev/test environemtn
if [ $ENV != "production" ]; then
    repoParent="/go/src/github.com/ncodes"
    mkdir -p $repoParent
    cd $repoParent
    git clone --depth=1 -b $VERSION https://github.com/ncodes/cocoon
    cd cocoon 
    git checkout $VERSION
    govendor fetch -v +out
    go build -o $GOPATH/bin/orderer core/orderer/main.go
else
     # Fetch pre-built binary 
    fileName="orderer_${VERSION}.zip"
    rm -rf $fileName
    rm -rf $GOPATH/bin/orderer
    printf "> Downloading pre-built binary [version: $VERSION, date: $(date)]\n"
    wget "https://storage.googleapis.com/hothot/${fileName}"
    unzip $fileName
    mv orderer $GOPATH/bin/orderer
    printf "Checksum: $(shasum $GOPATH/bin/orderer)\n"
fi

# start the orderer
$GOPATH/bin/orderer start