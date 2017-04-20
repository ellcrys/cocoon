bin-connector:
	@bash -c "APP=connector VERSION=$(version) bash './scripts/build.sh'"
bin-api:
	@bash -c "APP=api VERSION=$(version) bash './scripts/build.sh'"
bin-orderer:
	@bash -c "APP=orderer VERSION=$(version) bash './scripts/build.sh'"
bin-client:
	@bash -c "APP=client VERSION=$(version) bash './scripts/build.sh'"
test: test-api test-connector test-orderer test-client test-runtime-go test-common test-types
test-api:
	@bash -c "go test -v ./core/api/api/..."
test-connector:
	@bash -c "go test -v ./core/connector/connector/..."
test-orderer:
	@bash -c "go test -v ./core/orderer/orderer/..."
test-client:
	@bash -c "go test -v ./core/client/client/..."
test-runtime-go:
	@bash -c "go test -v ./core/runtime/golang/..."
test-common:
	@bash -c "go test -v ./core/common/..."
test-types:
	@bash -c "go test -v ./core/types/..."