# Launch Script

# start docker daemon
bash dockerd-entrypoint.sh dockerd &
printf "> Started docker daemon \n"
sleep 5

# pull launch-go image
docker pull ncodes/launch-go:latest

# pull cocoon source
repoOwner=github.com/ncodes
repoOwnerDir=$GOPATH/src/$repoOwner
mkdir -p $repoOwnerDir
cd $repoOwnerDir
branch="connector-redesign"
printf "> Fetching cocoon source. [branch=$branch] [dest=$repoOwnerDir]\n"
git clone --depth=1 -b $branch https://$repoOwner/cocoon

# build the binary
printf "> Building cocoon"
cd cocoon
rm -rf .glide/ && rm -rf vendor
glide --debug install
# go build -v -o /bin/cocoon

# start connector 
# printf "Running Cocoon Connector"
# cocoon connector
sleep 1d