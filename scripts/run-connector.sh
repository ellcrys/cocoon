# Run the connector 
set -e

# create a bridge 
# printf "> Creating bridge\n"
# bridge=""
# for (( ; ; ))
# do
#    bridge=cc_$(shuf -i 1-10000 -n 1)
#    printf "Create bridge [$bridge]"
#    brctl addbr $bridge
#    randIP=173.$(shuf -i 1-255 -n 1).$(shuf -i 0-255 -n 1).0
#    ip addr add $randIP dev $bridge
#    ip link set dev $bridge up
#    export BRIDGE_NAME=$bridge
#    break
# done
randIP=173.$(shuf -i 1-255 -n 1).$(shuf -i 0-255 -n 1).0

# start docker daemon
bash dockerd-entrypoint.sh dockerd --ip=$randIP &
printf "> Started docker daemon \n"
sleep 5

# pull launch-go image
docker pull ncodes/launch-go:latest

# pull cocoon source
branch="connector-redesign"
repoOwner=github.com/ncodes
repoOwnerDir=$GOPATH/src/$repoOwner
mkdir -p $repoOwnerDir
cd $repoOwnerDir
printf "> Fetching cocoon source. [branch=$branch] [dest=$repoOwnerDir]\n"
git clone --depth=1 -b $branch https://$repoOwner/cocoon

# build the binary
printf "> Building cocoon"
cd cocoon
rm -rf .glide/ && rm -rf vendor
glide --debug install
go build -v -o /bin/cocoon core/main.go

# start connector 
printf "Running Cocoon Connector"
cocoon connector