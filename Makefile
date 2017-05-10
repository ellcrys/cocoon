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
	@bash -c "protoc --proto_path=./vendor -I ./core/stub/proto_runtime/ ./core/stub/proto_runtime/server.proto --gogo_out=plugins=grpc:./core/stub/proto_runtime"
	@bash -c "echo Done!"
test: test-common test-types test-lock test-store test-bcm test-platform test-api test-connector test-orderer test-client test-stub
test-api:
	@bash -c "go test -v ./core/api/api/..."
test-connector:
	@bash -c "go test -v ./core/connector/..."
test-orderer:
	@bash -c "go test -v ./core/orderer/orderer/..."
test-client:
	@bash -c "go test -v ./core/client/client/..."
test-stub:
	@bash -c "go test -v ./core/stub/..."
test-common:
	@bash -c "go test -v ./core/common/..."
test-types:
	@bash -c "go test -v ./core/types/..."
test-lock:
	@bash -c "go test -v ./core/lock/..."
test-store:
	@bash -c "go test -v ./core/store/..."
test-bcm:
	@bash -c "go test -v ./core/blockchain/..."
test-platform:
	@bash -c "go test -v ./core/platform/..."
	