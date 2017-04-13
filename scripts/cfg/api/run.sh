# Run an API. 

repoParent="/go/src/github.com/ncodes"
mkdir -p $repoParent
cd $repoParent

# pull cocoon source
branch="master"
git clone --depth=1 -b $branch https://github.com/ncodes/cocoon
cd cocoon 
git checkout $branch

# start the orderer
glide --debug install
go build -o cocoon core/main.go
./cocoon api start