FROM ncodes/cocoon-go:latest

# add dependencies
RUN apt-get update
RUN apt-get -y install aufs-tools iptables git
RUN curl https://glide.sh/get | sh
RUN go get -u github.com/kardianos/govendor