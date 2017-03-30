# Launch Script

# start docker daemon
bash dockerd-entrypoint.sh dockerd &
printf "> Started docker daemon \n"
sleep 5

# pull launch-go image
docker pull ncodes/launch-go:latest

# pull cocoon source
cd /home
branch="connector-redesign"
printf "> Fetching cocoon source. [branch=$branch]\n"
git clone --depth=1 -b $branch https://github.com/ncodes/cocoon
cd cocoon/core

# build the binary
#glide --debug install
printf "> Sleep for an hour"
sleep 1h
# printf "Building cocoon source \n"
# go build -v -o /bin/cocoon

# start connector 
# repoHash=$(git rev-parse HEAD)
# printf "Cocoon Version: $repoHash \n"
# cocoon connector