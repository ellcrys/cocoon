# Cocoon - A no-gimmick, centralized, scalable smart contract platform. 

Cocoon is a smart contract engine that is designed to be fast, scalable and built on everyday technologies we use and 
love. It is centralized and includes a distributed, chained and ledger called a TransactionChain (TxChain) ledger. Our motives for building a centralized smart contract platform are as follows: 

- **Scalablity:**
 
Existing platforms like Ethereum are unable to handle high transactions per seconds. While the reasons are understood, this is unacceptable for building and hosting applications that will serve thounsands of concurrent user requests at any given time. This project is targetted towards projects where scabality is a priority and complete trustlessness is not a priority. 

- **Slow Transactions**

The blockchain is an important technology at the heart of decentralized systems such as Bitcoin and Ethereum. It allows transactions to be collected in series of blocks and these blocks are then broadcasted to every node and cryptographically linked to existing blocks in such a way that it is difficult to alter these data without reconstructing the links between the blocks on every single node on the network. While amazing, the block replication and replayed computation techniques makes the blockchain an inefficient datastore for centralized, high-throuhtput, permissioned applications like this project. In our opinion, the blockchain is a breakthrough technology designed to meet the security needs of decentralized systems and a complete adaptation to centralized, permissioned platform unnecessarily introduces foriegn performance issues.

**Expensive & Unterministic Cost**

Decentralized systems like Ethereum have opened our minds to the possibility of building open, autonomous applications
that require little or no human input to function. To ensure these open platforms are secured against attacks and spam, developers are required to pay fees for computations performed by their apps. The exact amount to pay is not immediately known and as such it makes it hard for developers and business to make decisions. 

We believe that smart contracts will become a huge part of how we interact with businesses and 
how businesses interact with each other and as such there needs to be a choice between going fully decentralized, trustless and accepting all the performance penalties that comes with it or building on a centralized system with the same functionalities (or even more), but with the benefit of fast, performant, synchronous transactions against a no-gimmick, immutable ledger (TxChain) and an insane ability to scale smart contracts vertically or horizontally. 

## Features
Below are the task list for features and capabilities we are looking to support. Only checked ones are currently functional. 

- [x] Distributed immutable ledger (BlocklessChain)
- [x] Smart contact engine (Initial support for Go programming language)
- [ ]  Horizontal/Vertical scalability of smart contracts. 
- [ ]  Ledger access control 
- [x] Support public ledger chain
- [ ]  Support private ledger chain
- [ ]  Multi-Sig smart contract deployment
- [ ]  Global lock API for smart contract
- [ ]  Transaction/Ledger streaming service
- [ ]  Multi-Sig transaction validation
- [ ]  Single Transaction API service
- [ ]  Cross contract messaging
- [ ]  Cross contract event service
- [ ]  Binary smart contract deployment
- [ ]  New contract language support for Node.js & Ruby
- [ ]  Native currency (Project Titan)

## Anatomy of Cocoon

***Note: This document is constantly being updated.***

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


