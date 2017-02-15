git clone https://github.com/ncodes/cocoon
git config --global http.https://gopkg.in.followRedirects true

# compilte data directory to binary
cd cocoon/core/data
go get -u github.com/jteeuwen/go-bindata/...
go-bindata --pkg data ./...

cd ../connector
glide install
go build -o /bin/connector connector.go