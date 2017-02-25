# Launch Script

# start docker daemon
dockerd-entrypoint.sh dockerd --iptables=false || { echo 'failed to start dockerd' ; exit 1; } & 
echo "Started docker daemon"

# pull cocoon source
git clone -b master https://github.com/ncodes/cocoon

# pull launch-go image
docker pull ncodes/launch-go:latest

# build the connector
cd cocoon/core/connector
glide install
go build -o /bin/connector connector.go

# start connector 
connector start