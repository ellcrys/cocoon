# Cocoon firewall directives
#TODO: find a way to delete this rules when the container shuts down

# #################### DOCKER BRIDGE SETUP ################
# Determine new bridge name. Use cocoon id as default bridge name
# but continously find a bridge name if the default has been used.
bridgeName="$(echo -n ${COCOON_ID} | awk '{print $1}' | cut -c1-15)"
for (( ; ; ))
do  
    brctl addbr ${bridgeName}
    if [ $? -eq 0 ]; then
        break
    fi
    randStr=$(cat /dev/urandom | tr -dc 'a-zA-Z0-9' | fold -w 32 | head -n 1)
    bridgeName="$(echo -n $randStr | sha1sum | awk '{print $1}' | cut -c1-15)"
done

# create new bridge
ip addr add dev $bridgeName
ip link set dev $bridgeName up
export BRIDGE_NAME=$bridgeName
echo "Create new bridge named: $bridgeName"

################### END DOCKER BRIDE SETUP #######################

# Drop all outgoing connections
iptables -C DOCKER -o $bridgeName -p tcp -j DROP
if [ $? -eq 1 ]; then
    iptables -I DOCKER 1 -o $bridgeName -p tcp -j DROP
    iptables -I DOCKER 1 -o $bridgeName -p udp -j DROP
fi

# set iptable rules
iptables -A FORWARD -o $bridgeName -j DOCKER
iptables -A FORWARD -o $bridgeName -m conntrack --ctstate RELATED,ESTABLISHED -j ACCEPT
iptables -A FORWARD -i $bridgeName ! -o $bridgeName -j ACCEPT
iptables -A FORWARD -i $bridgeName -o $bridgeName -j ACCEPT

# set postrouting rule
iface="$(docker network inspect bridge | jq '.[0].IPAM.Config[0].Subnet')"
iptables -t nat -A POSTROUTING -s $iface -j MASQUERADE