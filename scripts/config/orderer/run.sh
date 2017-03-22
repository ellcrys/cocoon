# Run an orderer
repoParent = "/go/src/github.com/ncodes"
mkdir -p $repoParent
cd $repoParent

# pull cocoon source
git clone --depth=1 https://github.com/ncodes/cocoon

# start the orderer
cd cocoon
go run core/main.go
