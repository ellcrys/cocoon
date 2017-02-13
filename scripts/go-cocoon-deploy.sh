# Setup script for go cocoon
# Expects to be called in /gopath/src/github.com/ncodes
cd cocoon/core

# build binary data files
go get -u github.com/jteeuwen/go-bindata/...
go-bindata --pkg data -o ./data/bindata.go ./data/...

# build connector and move binary to path
cd connector
glide update
go build -o /bin/connector connector.go