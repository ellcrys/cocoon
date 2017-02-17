# pull cocoon source
git clone -b master https://github.com/ncodes/cocoon
git config --global http.https://gopkg.in.followRedirects true

# compilte data directory to binary
cd cocoon/core/data
go get -u -v github.com/jteeuwen/go-bindata/...
go-bindata --pkg data ./...

# build connector
cd ../connector
glide install
go build -o /bin/connector connector.go

# add firewall
sudo groupadd cocoon
sudo iptables -A OUTPUT -o lo -p tcp -m owner --gid-owner cocoon -j DROP
sudo iptables -A OUTPUT -o lo -p udp -m owner --gid-owner cocoon -j DROP

# start connector
connector start