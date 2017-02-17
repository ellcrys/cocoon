# Script to setup the cocoon docker image. (privlledged flag is required).
# This file must be executed during image building phase. 

iptables -A INPUT -m conntrack --ctstate ESTABLISHED,RELATED -j ACCEPT

# Default cocoon code group prevents access to connector server,
# public internet
groupdel cocoon && groupadd cocoon
iptables -A OUTPUT -o lo -p tcp --dport 3000 -m owner --gid-owner cocoon -j DROP
iptables -A OUTPUT -o lo -p udp --dport 3000 -m owner --gid-owner cocoon -j DROP
iptables -A OUTPUT -o lo -p tcp --dport 3001 -m owner --gid-owner cocoon -j ACCEPT
iptables -A OUTPUT -m owner --gid-owner cocoon -j DROP

# This cocoon group allows access to the public internet
groupdel cocoon-open && groupdel cocoon-open
iptables -A OUTPUT -o lo -p tcp --dport 3000 -m owner --gid-owner cocoon-open -j DROP
iptables -A OUTPUT -o lo -p udp --dport 3000 -m owner --gid-owner cocoon-open -j DROP
iptables -A OUTPUT -o lo -p tcp --dport 3001 -m owner --gid-owner cocoon-open -j ACCEPT

# save
iptables-save > /etc/iptables/rules.v4