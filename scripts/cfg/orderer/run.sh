# Run an orderer. 
# Expects the following variables to be use to configure the orderer. This should
# typically be set by a scheduler.
# STORE_CON_STR=host=localhost user=ned dbname=cocoon sslmode=disable password=
pwd
repoParent="/go/src/github.com/ncodes"
mkdir -p $repoParent
cd $repoParent

# pull cocoon source
branch="master"
git clone --depth=1 -b $branch https://github.com/ncodes/cocoon
cd cocoon 
git checkout $branch

# start the orderer
govendor fetch -v +out
go build -o orderer core/orderer/main.go
./orderer start