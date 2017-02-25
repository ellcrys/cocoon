# Launch Script

# start docker daemon
bash dockerd-entrypoint.sh dockerd --iptables=false &
echo "Started docker daemon"
sleep 5

# pull launch-go image
docker pull ncodes/launch-go:latest

# pull cocoon source
git clone -b master https://github.com/ncodes/cocoon

# build the binary
cd cocoon
glide install
echo "Building cocoon soruce"
go build -o /bin/cocoon core/main.go 

# start connector 
repoHash=$(git rev-parse HEAD)
echo "Cocoon Version: $repoHash"
cocoon connector