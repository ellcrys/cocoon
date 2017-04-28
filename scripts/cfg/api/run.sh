# Run an API. 
set -e 

VERSION=$API_VERSION

# Fetch and build API from source in dev/test environemtn
if [ $ENV != "production" ]; then
    repoParent="/go/src/github.com/ncodes"
    mkdir -p $repoParent
    cd $repoParent
    git clone --depth=1 -b $VERSION https://github.com/ncodes/cocoon
    cd cocoon 
    git checkout $VERSION
    govendor fetch -v +out
    go build -o $GOPATH/bin/api core/api/main.go
else
     # Fetch pre-built binary 
    fileName="api_${VERSION}.zip"
    rm -rf $fileName
    rm -rf $GOPATH/bin/api
    printf "> Downloading pre-built binary [version: $VERSION, date: $(date)]\n"
    wget "https://storage.googleapis.com/hothot/${fileName}"
    unzip $fileName
    mv api $GOPATH/bin/api
    printf "Checksum: $(shasum $GOPATH/bin/api)\n"
fi

# start the api
exec $GOPATH/bin/api start