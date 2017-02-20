# Launch Script

# #################### Docker Bridge Setup ################
# ip addr show docker0
# if [ $? -eq 0 ]; then
    # ip link set dev docker0 down
    #brctl delbr docker0
    # iptables -t nat -F POSTROUTING
# fi

# Create bridge. Attempt to use the cocoon id as the bridge bridgeName
# otherwise, continusly create new names till we find a unique one
bridgeName=${COCOON_ID}
for (( ; ; ))
do  
    brctl addbr ${bridgeName}
    if [ $? -eq 0 ]; then
        break
    fi
    randStr=$(cat /dev/urandom | tr -dc 'a-zA-Z0-9' | fold -w 32 | head -n 1)
    bridgeName="$(echo -n $randStr | sha1sum | awk '{print $1}' | cut -c1-15)"
done

echo "Bridge Name: $bridgeName"

export BRIDGE_NAME=$bridgeName

ip addr add dev $bridgeName
ip link set dev $bridgeName up

docker-entrypoint.sh dockerd -b=$bridgeName --iptables=false & 
################### END DOCKER BRIDE SETUP #######################

# give dockerd time to start
sleep 10

##################### SET COCOON FIREWALL AND GROUPS #############
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