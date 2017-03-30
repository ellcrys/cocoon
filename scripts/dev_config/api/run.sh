# Run an API. 

repoParent="/go/src/github.com/ncodes"
mkdir -p $repoParent
cd $repoParent

# pull cocoon source
branch="connector-redesign"
git clone --depth=1 -b $branch https://github.com/ncodes/cocoon
cd cocoon 
git checkout $branch

# start the orderer
cd core
glide --debug update
go build -o cocoon
./cocoon api start