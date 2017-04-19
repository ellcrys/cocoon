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
    printf "> Downloading pre-built binary [version: $VERSION]\n"
    wget "https://storage.googleapis.com/krogan/orderer_${VERSION}.zip"
    unzip "orderer_${VERSION}.zip"
    mv orderer $GOPATH/bin/orderer
fi

# start the orderer
$GOPATH/bin/orderer start