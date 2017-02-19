# Cocoon firewall directives

bridgeName=${BRIDGE_NAME}

# Drop all outgoing connections
iptables -C DOCKER -o $bridgeName -p tcp -j DROP
if [ $? -eq 1 ]; then
    iptables -I DOCKER 1 -o $bridgeName -p tcp -j DROP
    iptables -I DOCKER 1 -o $bridgeName -p udp -j DROP
fi