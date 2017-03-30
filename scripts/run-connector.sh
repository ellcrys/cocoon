#!/bin/bash
# Run the connector 
set -e

trap "echo Booh!" SIGINT SIGTERM
echo "pid is $$"

while :                 # This is the same as "while true".
do
    a=1
done

# create a bridge 
printf "> Creating bridge\n"
bridge=""
for (( ; ; ))
do
   # if bridge name is taken, 'continue'' loop
   bridge=cc_$(shuf -i 1-10000 -n 1)
   if ip addr | grep $bridge; then
        continue
   fi
   
   # if bridge ip is taken, 'continue' loop
   bridgeIP=173.$(shuf -i 18-255 -n 1).0.$(shuf -i 0-255 -n 1)
   if ip addr | grep $bridgeIP; then 
        continue
   fi
   
   brctl addbr $bridge
   ip addr add $bridgeIP/16 dev $bridge
   ip link set dev $bridge up
   
   export BRIDGE_NAME=$bridge
   export BRIDGE_IP=bridgeIP
   printf '> Created bridge \n'
   break
done

# start docker daemon
bash dockerd-entrypoint.sh dockerd --bridge=$bridge &
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

# delete bridge 
trap 'ip link set dev $bridge down; brctl delbr $bridge' TERM INT

