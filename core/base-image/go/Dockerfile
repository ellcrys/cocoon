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
RUN git config --global http.https://gopkg.in.followRedirects true
RUN git clone https://github.com/ncodes/cocoon -b master

# run go cocoon deployment script
CMD chmod +x cocoon/script/go-cocoon-deploy.sh && ./cocoon/script/go-cocoon-deploy.sh