# Launch Script

# start docker daemon
bash dockerd-entrypoint.sh dockerd &
printf "Started docker daemon \n"
sleep 5

# pull launch-go image
docker pull ncodes/launch-go:latest

# pull cocoon source
branch="connector-redesign"
git clone --depth=1 -b $branch https://github.com/ncodes/cocoon
cd cocoon 
git checkout $branch

# build the binary
cd core
glide --debug update
printf "Building cocoon source \n"
go build -v -o /bin/cocoon

# start connector 
repoHash=$(git rev-parse HEAD)
printf "Cocoon Version: $repoHash \n"
cocoon connector