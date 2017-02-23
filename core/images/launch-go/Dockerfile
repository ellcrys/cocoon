FROM golang:1.7.5-wheezy

# add dependencies
RUN apt-get update
RUN apt-get install aufs-tools
RUN curl https://glide.sh/get | sh

# for testing
RUN go get -v github.com/jpillora/go-echo-server