# Run an orderer
repoParent = "/go/src/github.com/ncodes"
mkdir $repoParent
cd $repoParent

# pull cocoon source
git clone --depth=1 https://github.com/ncodes/cocoon

# start the orderer
go run core/main.go
