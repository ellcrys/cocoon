cd ../core/connector

export COCOON_ID=abc1
export COCOON_CODE_URL=https://github.com/ncodes/cocoon-example-01
export COCOON_CODE_LANG=go
export COCOON_BUILD_PARAMS='{ "pkg_mgr": "glide" }'

go run connector.go start