FROM ncodes/cocoon:latest

# add dependencies
RUN apk update
RUN apk add git bash glide shadow jq

# create working directorys
WORKDIR src
WORKDIR github.com
WORKDIR ncodes