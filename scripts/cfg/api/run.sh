# Run an API. 

repoParent="/go/src/github.com/ncodes"
mkdir -p $repoParent
cd $repoParent

# pull cocoon source
branch="master"
git clone --depth=1 -b $branch https://github.com/ncodes/cocoon
cd cocoon 
git checkout $branch

# start the api
glide install
go build -o api core/api/main.go
./api start