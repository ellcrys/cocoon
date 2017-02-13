FROM golang:1.7.5-alpine3.5

RUN apk update
RUN apk add bash
RUN apk add nano
RUN apk add git
RUN apk add glide

# create working directors
WORKDIR src
WORKDIR github.com
WORKDIR ncodes