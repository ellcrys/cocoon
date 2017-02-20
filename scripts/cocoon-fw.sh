# Cocoon firewall directives
#TODO: find a way to delete this rules when the container shuts down

bridgeName=${BRIDGE_NAME}

# Drop all outgoing connections
# iptables -C DOCKER -o $bridgeName -p tcp -j DROP
# if [ $? -eq 1 ]; then
    # iptables -I DOCKER 1 -o $bridgeName -p tcp -j DROP
    # iptables -I DOCKER 1 -o $bridgeName -p udp -j DROP
# fi

# set forwarding rules. 
iptables -A FORWARD -o $bridgeName -j DOCKER
iptables -A FORWARD -o $bridgeName -m conntrack --ctstate RELATED,ESTABLISHED -j ACCEPT
iptables -A FORWARD -i $bridgeName ! -o $bridgeName -j ACCEPT
iptables -A FORWARD -i $bridgeName -o $bridgeName -j ACCEPT