#!/usr/bin/env bash
# This scripts the various applications from source
if [ $(uname) != "Linux" ]; then 
    echo "Unable to build on $(uname). Use Linux."
    exit
fi

version=""
if [ $VERSION != "" ]; then 
    version=$VERSION
fi

# change client name to the official plaform name 
appName=${APP}
if [ $appName == "client" ]; then 
    appName="ellcrys"
fi

# remove and create build dir
rm -rf pkg/*
mkdir -p pkg

# Build the app
echo "==> Building $APP for linux amd64"
CGO_ENABLED=1 GOARCH="amd64" GOOS="linux" go build -o "pkg/linux_amd64/${appName}" core/$APP/main.go

# zip app
cd pkg/linux_amd64
zip "${APP}_${version}.zip" ${appName}
rm -rf ${appName}