# Cocoon firewall directives

# Drop all outgoing connections
iptables -C DOCKER -o docker0 -p tcp -j DROP
if [ $? -eq 1 ]; then
    iptables -I DOCKER 1 -o docker0 -p tcp -j DROP
    iptables -I DOCKER 1 -o docker0 -p udp -j DROP
else
    iptables -D DOCKER -o docker0 -p tcp -j DROP
    iptables -D DOCKER -o docker0 -p udp -j DROP
fi
