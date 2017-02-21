# Launch Script

# Run cocoon firewall script
bash ${NOMAD_META_SCRIPTS_DIR}/${NOMAD_META_COCOON_FIREWALL_SCRIPT_NAME}

# pull cocoon source
git clone -b master https://github.com/ncodes/cocoon
git config --global http.https://gopkg.in.followRedirects true

# compilte data directory to binary
cd cocoon/core/data
go get -u -v github.com/jteeuwen/go-bindata/...
go-bindata --pkg data ./...

# pull launch-go image
docker pull ncodes/launch-go:latest

# build the connector
cd ../connector
glide install
go build -o /bin/connector connector.go

# start connector 
connector start