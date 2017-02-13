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

# clone cocoon source & follow redirects for gopkg
RUN git clone https://github.com/ncodes/cocoon
RUN git config --global http.https://gopkg.in.followRedirects true

# compilte data directory to binary
WORKDIR cocoon/core/data
RUN go get -u github.com/jteeuwen/go-bindata/...
RUN go-bindata --pkg data ./...

WORKDIR ../connector
RUN glide update
RUN go build -o /bin/connector connector.go