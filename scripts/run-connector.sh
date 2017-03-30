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
printf "> Building cocoon"
rm -rf .glide/ && rm -rf vendor
glide --debug install
go build -v -o /bin/cocoon

# start connector 
printf "Running Cocoon Connector"
cocoon connector