# pull cocoon source
git clone -b master https://github.com/ncodes/cocoon
git config --global http.https://gopkg.in.followRedirects true

# compilte data directory to binary
cd cocoon/core/data
go get -u -v github.com/jteeuwen/go-bindata/...
go-bindata --pkg data ./...

# build connector
cd ../connector
glide install
cat ccode/ccode.go
go build -o /bin/connector connector.go

# start connector
connector start