# Launch Script

# start docker daemon
bash dockerd-entrypoint.sh dockerd --iptables=false || { echo 'failed to start dockerd' ; exit 1; } & 
echo "Started docker daemon"

# pull cocoon source
git clone -b master https://github.com/ncodes/cocoon || { echo 'failed to pull cocoon repository' ; exit 1; }

# pull launch-go image
docker pull ncodes/launch-go:latest || { echo 'failed to pull ncodes/launch-go image' ; exit 1; }

# build the connector
cd cocoon/core/connector
glide install || { echo 'glide install has failed' ; exit 1; }
go build -o /bin/connector connector.go || { echo 'failed to build connector' ; exit 1; }

# start connector 
connector start