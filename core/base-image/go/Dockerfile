FROM golang:1.7.5-alpine3.5

RUN apk update
RUN apk add bash
RUN apk add nano
RUN apk add git
RUN apk add glide

# install cocoon source code
WORKDIR src
WORKDIR github.com
WORKDIR ncodes
RUN git clone https://github.com/ncodes/cocoon -b master

# install & start the connector
WORKDIR cocoon/core/connector
RUN glide update
RUN go build -o /bin/connector connector.go
CMD connector start