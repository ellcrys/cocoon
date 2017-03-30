# Run the connector 
set -e

# create a bridge 
printf "> Creating bridge\n"
bridge=""
for (( ; ; ))
do
   bridge=cc_$(shuf -i 1-10000 -n 1)
   printf "Create bridge [$bridge]"
   brctl addbr $bridge
   ip=$(nmap -n -iR 1 --exclude 10.0.0.0/8,127.0.0.0/8,172.16.0.0/32,192.168.0.0/16,224-255.-.-.- -sL | awk 'FNR==3{print $5  }')
   ip addr add $ip dev $bridge
   ip link set dev $bridge up
   export BRIDGE_NAME=$bridge
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