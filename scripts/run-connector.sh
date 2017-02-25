# Launch Script

# start docker daemon
bash dockerd-entrypoint.sh dockerd --iptables=false || { echo 'failed to start dockerd' ; exit 1; } & 
echo "Started docker daemon"

# pull launch-go image
docker pull ncodes/launch-go:latest || { echo 'failed to pull ncodes/launch-go image' ; exit 1; }

# pull cocoon source
git clone -b master https://github.com/ncodes/cocoon || { echo 'failed to pull cocoon repository' ; exit 1; }

# build the binary
cd cocoon
glide install || { echo 'glide install has failed' ; exit 1; }
go build -o /bin/cocoon core/main.go || { echo 'failed to build cocoon binary' ; exit 1; }

# start connector 
repoHash=$(git rev-parse HEAD)
echo "Cocoon Version: $repoHash"
cocoon connector