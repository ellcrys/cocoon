FROM golang:1.7.5-alpine3.5

RUN apk update
RUN apk add bash
RUN apk add nano
RUN apk add git

# install cocoon source code
WORKDIR src
WORKDIR github.com
WORKDIR ncodes
RUN git clone https://github.com/ncodes/cocoon 
WORKDIR cocoon