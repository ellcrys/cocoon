FROM golang:1.7.5-wheezy

RUN apt-get update
RUN apt-get install -y apt-utils
RUN apt-get -y install sudo
RUN apt-get -y install iptables
RUN apt-get -y install iptables-persistent

# create working directors
WORKDIR src
WORKDIR github.com
WORKDIR ncodes