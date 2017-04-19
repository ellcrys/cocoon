# Run an API. 
set -e 

version="master"

# Fetch and build API from source in dev/test environemtn
if [ $ENV != "production" ]; then
    repoParent="/go/src/github.com/ncodes"
    mkdir -p $repoParent
    cd $repoParent
    git clone --depth=1 -b $version https://github.com/ncodes/cocoon
    cd cocoon 
    git checkout $version
    govendor fetch -v +out
    go build -o $GOPATH/bin/api core/api/main.go
else
     # Fetch pre-built binary 
    printf "> Downloading pre-built binary [version: $VERSION]\n"
    wget "https://storage.googleapis.com/krogan/api_${VERSION}.zip"
    unzip "api_${VERSION}.zip"
    mv api $GOPATH/bin/api
fi

# start the api
$GOPATH/bin/api start