# Launch Script

# start docker daemon
bash dockerd-entrypoint.sh dockerd &
printf "Started docker daemon \n"
sleep 5

# pull launch-go image
docker pull ncodes/launch-go:latest

# pull cocoon source
git clone --depth=1 -b=connector-redesign https://github.com/ncodes/cocoon

# build the binary
cd cocoon
glide install
printf "Building cocoon source \n"
go build -v -o /bin/cocoon core/main.go 

# start connector 
repoHash=$(git rev-parse HEAD)
printf "Cocoon Version: $repoHash \n"
cocoon connector