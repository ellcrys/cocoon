# Script to setup the cocoon docker image. (privlledged flag is required)

iptables -A INPUT -m conntrack --ctstate ESTABLISHED,RELATED -j ACCEPT

# Default cocoon code group prevents access to connector server,
# public internet
groupadd cocoon
iptables -A OUTPUT -o lo -dport 3000 -m owner --gid-owner cocoon -j DROP
iptables -A OUTPUT -o lo -dport 3001 -m owner --gid-owner cocoon -j ACCEPT
iptables -A OUTPUT -m owner --gid-owner cocoon -j DROP

# This cocoon group allows access to the public internet
groupadd cocoon-open
iptables -A OUTPUT -o lo -dport 3000 -m owner --gid-owner cocoon-open -j DROP
iptables -A OUTPUT -o lo -dport 3001 -m owner --gid-owner cocoon-open -j ACCEPT
