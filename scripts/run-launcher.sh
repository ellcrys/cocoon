# setup docker.
# Create custom bridge
ip link set dev docker0 down
brctl delbr docker0
iptables -t nat -F POSTROUTING

# Create bridge. Attempt to use the cocoon id as the bridge bridgeName
# otherwise, continusly create new names till we find a unique one
bridgeName=${COCOON_ID}
for (( ; ; ))
do  
    brctl addbr bridgeName
    if [$? -eq 0]
        break
    fi
    randStr=$(cat /dev/urandom | tr -dc 'a-zA-Z0-9' | fold -w 32 | head -n 1)
    bridgeName="$(echo -n $randStr | sha1sum | awk '{print $1}' | cut -c1-15)"
done

ip addr add 192.168.5.1/24 dev $bridgeName
sudo ip link set dev $bridgeName up

docker-entrypoint.sh dockerd -b=$bridgeName &

# give dockerd time to start
sleep 10

# set coccon firewall and groups
bash ${NOMAD_META_SCRIPTS_DIR}/${NOMAD_META_COCOON_GROUPS_SCRIPT_NAME}

# pull cocoon source
git clone -b master https://github.com/ncodes/cocoon
git config --global http.https://gopkg.in.followRedirects true

# compilte data directory to binary
cd cocoon/core/data
go get -u -v github.com/jteeuwen/go-bindata/...
go-bindata --pkg data ./...

# pull launch-go image
docker pull ncodes/launch-go:latest

# launcher launcher
cd ../launcher
glide install
go build -o /bin/launcher launcher.go

# start launcher
launcher start