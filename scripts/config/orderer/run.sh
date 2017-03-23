# Run an orderer. 
# Expects the following variables to be use to configure the orderer. This should
# typically be set by a scheduler.
# ORDERER_ADDR=127.0.0.1:8001
# STORE_CON_STR=host=localhost user=ned dbname=cocoon sslmode=disable password=

repoParent="/go/src/github.com/ncodes"
mkdir -p $repoParent
cd $repoParent

# pull cocoon source
git clone --depth=1 https://github.com/ncodes/cocoon

# start the orderer
cd cocoon
glide install
go run core/main.go orderer
