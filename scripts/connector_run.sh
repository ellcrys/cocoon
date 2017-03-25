export COCOON_ID=abc1
export COCOON_CODE_URL=https://github.com/ncodes/cocoon-example-01
export COCOON_CODE_LANG=go
export COCOON_BUILD_PARAMS='eyAicGtnX21nciI6ICJnbGlkZSIgfQ=='
#export DEV_ORDERER_ADDR=       # directly set the orderer addr 
#export DEV_RUN_ROOT_BIN=       # force launcher to ignore running the cocoon code build routine and just run a `ccode` binary in the cocoon code source root
#export DEV_COCOON_ADDR=        # directly set the address to the cocoon code server the connector client connects to. 
#export COCOON_DISK_LIMIT=      # set the disk limit of the cocoon code
go run core/main.go connector