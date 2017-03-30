FROM ncodes/cocoon:latest

# add dependencies
#RUN echo "deb http://http.debian.net/debian wheezy-backports main" >> /etc/apt/sources.list
RUN apt-get update
RUN apt-get -y install aufs-tools bridge-utils nmap iproute2
RUN curl https://glide.sh/get | sh

# create working directorys
WORKDIR src
WORKDIR github.com
WORKDIR ncodes