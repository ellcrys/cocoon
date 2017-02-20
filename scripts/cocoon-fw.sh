# Cocoon firewall directives
#TODO: find a way to delete this rules when the container shuts down

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

export BRIDGE_NAME=$bridgeName

# create new bridge
ip addr add dev $bridgeName
ip link set dev $bridgeName up
echo "Create new bridge named: $bridgeName"


# start docker daemon
docker-entrypoint.sh dockerd -b=$bridgeName --iptables=false & 
sleep 5
echo "Started docker daemon"

# forward bride packets to DOCKER chain
iptables -A FORWARD -o $bridgeName -j DOCKER
iptables -A FORWARD -o $bridgeName -m conntrack --ctstate RELATED,ESTABLISHED -j ACCEPT

# Accept DNS lookup packets
dnsIPs="$(cat /etc/resolv.conf | grep 'nameserver' | cut -c12-)"
for ip in $dnsIPs
do
    iptables -A FORWARD -m state --state NEW,ESTABLISHED -d ${ip} -p udp --dport 53  -i $bridgeName -j ACCEPT
    iptables -A FORWARD -m state --state ESTABLISHED -p udp -s ${ip} --sport 53 -o $bridgeName -j ACCEPT
    iptables -A FORWARD -m state --state NEW,ESTABLISHED -d ${ip} -p tcp --dport 53  -i $bridgeName -j ACCEPT
    iptables -A FORWARD -m state --state ESTABLISHED -p tcp -s ${ip} --sport 53  -o $bridgeName -j ACCEPT
done

# Block all udp and tcp packets
iptables -A FORWARD -p udp  -i $bridgeName -j DROP
iptables -A FORWARD -p tcp -i $bridgeName -j DROP

# set container forwarding rules
iptables -A FORWARD -i $bridgeName ! -o $bridgeName -j ACCEPT           # route tracfic from bride to any other available bridge (other containers & internet)
iptables -A FORWARD -i $bridgeName -o $bridgeName -j ACCEPT

# set postrouting rule
iface="$(docker network inspect bridge | jq --raw-output '.[0].IPAM.Config[0].Subnet')"
iptables -t nat -A POSTROUTING -s $iface -j MASQUERADE