# Launch Script

# start docker daemon
dockerd-entrypoint.sh dockerd --iptables=false || { echo 'failed to start dockerd' ; exit 1; } & 
echo "Started docker daemon"

# pull cocoon source
git clone -b master https://github.com/ncodes/cocoon

# compile data directory to binary
cd cocoon/core/data
go get -u -v github.com/jteeuwen/go-bindata/...
go-bindata --pkg data ./...

# pull launch-go image
docker pull ncodes/launch-go:latest

# build the connector
cd ../connector
glide install
go build -o /bin/connector connector.go

# start connector 
connector start