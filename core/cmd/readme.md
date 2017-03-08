## Introduction To Cocoon Services

This document describes the services available and how to start/run them on a development machine. The contents of this document is targeted towards people who are interested in knowing how the system works, looking to contribute to the project or just looking to give it a spin. 

These are the available services on the Cocoon platform:

1. Connector
2. Orderer
3. Stub (CocoonCode a.k.a Smart Contract)
4. API 
5. Client 


These are commands for starting various complementary services such as the connector, orderer etc on a development machine. This document also contains the required and optional environment variables for each command. Because our scheduler is Nomad, some of the environment variables are prefixed with `NOMAD_` keyword.

## 1. Connector
The connector is a service that starts a cocoon code, monitors and manages it. It is the interface between the every other services and the cocoon code and as such the cocoon code can only communicate with other services through the connector. It works by establishing a connection to the cocoon code as soon as it is started and relays every relevant instructions to and from the cocoon code.  

##### Command
```sh
go run core/main.go connector
```

##### Environment variables

| Environment Variable| Required      | Default Value  | Description  | 
| --------------------|:-------------| --------------|:------------|
| NOMAD_IP_connector  | true          |                | The IP address to connect the connnectors GRPC server to. 
| NOMAD_PORT_connector| true          |                | The GRPC server port to bind to.
| COCOON_CODE_PORT    | true          |     8000       | The GRPC server port to bind to.
| DEV_COCOON_CODE_PORT| false         |     8003       | The port to a local cocoon code running on a separate process on the machine. If this is set, the connector will not attempt to launch a cocoon code. It will establish a connection with the cocoon code at this port.
| DEV_ORDERER_ADDR    | false          |                | The port to a local orderer service. A cocoon code will not be able to do anything meaningful with out the connector having access to an orderer.
| HOME                | true           |    $HOME       | The home directory of the machine. Required by the Go language deployment implementation. 
| COCOON_ID           | true          |                 | A unique id for the CocoonCode 
| COCOON_CODE_URL     | true          |                 | A github link to the CocoonCode source code 
| COCOON_CODE_TAG     | false         | latest          | The github release tag to fetch and run
| COCOON_DISK_LIMIT   | false         | 300MB           | The amount of ephemeral disk space a cocoon can use before it is restarted
| COCOON_BUILD_PARAMS | false         |                 | Standard Base64 encoded configuration options for building a cocoon code. e.g `Base64Encode("{\"pkg_mgr\": \"glide\"}")`

## 2. Orderer
The orderer is the gateway service to the immutable ledger shared by every smart contract. It runs an implementation
of the ledger interface based on our simple, blockless chain design. Below is a basic description of the blockless chain.  

### 2.1. BlocklessChain - A no gimmick, immutable, chained transactions built on proven technologies. 

A blockless chain is a ridiculously simple database structure that collects transactions and cryptograhically links them together. Each transaction referencing the hash before it and as such a change to a transaction effectively invalidates the entire chain. It is built on existing, proven technologies with great replication techniques. Our current implementation is based on Postgres and supports easy plugging of other implementations. The blockless chain allows smart contracts to create as many chains/ledgers as they want and have the option to make them publicly accessible or private. By default, all smart contracts have access to the global chain which is a publicly accessible. 

##### Command
```sh
go run core/main.go orderer
```

##### Environment variables

| Environment Variable          | Required      | Default Value  | Description  | 
| ------------------------------|:--------------| ---------------|:-------------|
| ORDERER_ADDR                  | true          | 127.0.0.1:8001 | The address to bind the orderer server to
| LEDGER_CHAIN_CONNECTION_STRING| true          |                | Connection string to postgres server


## 3. Stub

The stub is a service that runs within the container (a.k.a cocoon) where the smart contract is placed. It is primary
communication interface between the smart contract code and the external platform services. It provides an API to access the global ledger or any ledger created by the smart contract. The service is typically started within the smart code implementation. 

## 4. API

The API services is the primary interface for accessing the platform in a production cluster. Clients can make requests to it to deploy or manage cocoon codes, identities, ledger chains etc. 

##### Command
```sh
go run core/api/main.go start  

# Flags
# - bind-addr (default: 8004)
# - scheduler-addr 
# - scheduler-addr-https (default:false)
```

##### Environment variables

| Environment Variable          | Required      | Default Value  | Description  | 
| ------------------------------|:--------------| ---------------|:-------------|
| SCHEDULER_ADDR                | true          |                | The scheduler address in the cluster

## 5. Client

The client is a command line utility for sending instructions to a cocoon platform. It allows a user to perform any operation from a terminal. Operations such as cocoon deployment, deployment authorization, cocoon resource management etc.


##### Command
```sh
go run core/client/main.go 
```

##### Environment variables

| Environment Variable          | Required      | Default Value  | Description  | 
| ------------------------------|:--------------| ---------------|:-------------|
| API_ADDRESS                   | true          | 127.0.0.1:8004 | The API service address
