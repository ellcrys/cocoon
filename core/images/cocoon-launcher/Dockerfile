FROM ellcrys/cocoon:latest

# add dependencies
#RUN echo "deb http://http.debian.net/debian wheezy-backports main" >> /etc/apt/sources.list
RUN apt-get update
RUN apt-get -y install aufs-tools bridge-utils nmap iproute2
RUN go get -u github.com/kardianos/govendor

# create working directorys
WORKDIR src
WORKDIR github.com
WORKDIR ellcrys