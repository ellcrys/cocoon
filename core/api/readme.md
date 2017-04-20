## Platform API

The platform API provides a GRPC service for creating and managing identities, cocoons, releases and other resources. 

### Start the service:
```sh
env SCHEDULER_ADDR=104.199.69.200:4646 \ 
    CONSUL_ADDR=104.199.69.200:8500 \
    DEV_ORDERER_ADDR=127.0.0.1:8001 \
    ENV=production \
    API_VERSION=1.0.0 \
    CONNECTOR_VERSION=1.0.0 \
    go run main.go start
```

### Environment Variables
- SCHEDULER_ADDR: The address to the scheduler 
- CONSUL_ADDR: The address to the consul server
- DEV_ORDERER_ADDR: (Optional) The address to an orderer service. If not specified, the platform API will attempt to discover it through consul.
- ENV (default: development): Setting this to `production` will fetch and run a pre-built binary of the API as opposed to building from source. This also applies to the connector associated with a cocoon. 
- API_VERSION: This is the version of the API to run. 
- CONNECTOR_VERSION: This is the version of the connector to run.
