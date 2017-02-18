# Cocoon firewall directives

# Drop all outgoing connections
iptables -I DOCKER 1 -o docker0 -p tcp -j DROP
iptables -I DOCKER 1 -o docker0 -p udp -j DROP