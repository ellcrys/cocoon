FROM ncodes/cocoon-go:dind

# add dependencies
RUN apk update
RUN apk add git bash glide

# create working directorys
WORKDIR src
WORKDIR github.com
WORKDIR ncodes