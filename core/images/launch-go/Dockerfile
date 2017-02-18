FROM golang:1.8.0-alpine

# add dependencies
RUN apk update
RUN apk add git bash glide

# for testing
RUN go get -v github.com/jpillora/go-echo-server