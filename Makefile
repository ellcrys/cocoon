bin-connector:
	@bash -c "APP=connector VERSION=$(version) bash './scripts/build.sh'"
bin-api:
	@bash -c "APP=api VERSION=$(version) bash './scripts/build.sh'"
bin-orderer:
	@bash -c "APP=orderer VERSION=$(version) bash './scripts/build.sh'"
bin-client:
	@bash -c "APP=client VERSION=$(version) bash './scripts/build.sh'"
gen-pb:
	@bash -c "protoc --proto_path=./vendor -I ./core/api/api/proto_api/ ./core/api/api/proto_api/server.proto --gogo_out=plugins=grpc:./core/api/api/proto_api"
	@bash -c "protoc --proto_path=./vendor -I ./core/connector/server/proto_connector/ ./core/connector/server/proto_connector/server.proto --gogo_out=plugins=grpc:./core/connector/server/proto_connector"
	@bash -c "protoc --proto_path=./vendor -I ./core/orderer/proto_orderer/ ./core/orderer/proto_orderer/server.proto --gogo_out=plugins=grpc:./core/orderer/proto_orderer"
	@bash -c "protoc --proto_path=./vendor -I ./core/runtime/golang/proto_runtime/ ./core/runtime/golang/proto_runtime/server.proto --gogo_out=plugins=grpc:./core/runtime/golang/proto_runtime"
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