### Cocoon 

Cocoon is an opinionated smart contract or self-executing application executor. It is the engine that powers
Ellcrys smart contract platform providing the ability to build, run and interact with isolated, multi-signature
 and immutable applications. Cocoon takes advantage of existing technologies such as Docker, Nomad, Postgres and more
to provide a fast, scalable and easy to use smart contract platform.

#### Is Cocoon another Ethereum?

Cocoon and Ethereum share a common goal to build a platform that can run applications autonomously. However, 
Ethereum takes the idea further by decentralizing the execution of these applications. This is great 
and can leading to a world where applications can run without censorship. Although, Ethereum is being actively developed, 
the project is proving to be hard to build and scale. Cocoon is the no-gimmick smart contract platform that leverages on existing, 
trusted tools and technologies to build an engine that can perform the function of executing a multi-signature/immutable application. Cocoon
is being built to run on centralized infrastructure.

#### Does Cocoon use a blockchain

No. Cocoon does not directly provide a blockchain. The primary data store for self-executing application to store state
is powered by an ACID-complaint, immutable database service. Immutability is provided as a feature (no delete operations).
We also have plans to support blockchain immutability by providing contract operations to store objects on the Ethereum 
blockchain.

#### What languages are supported?

Contract applications are currently built with the Go programming language. We are working
to support Javascript, Ruby, PHP and Python soon.


### Features 

- **Multi-Signature Support**: Build contract applications owned and controlled by more that one entity.
- **Immutable Contracts**: Contract applications are cloned from Github, versioned and archived forever. 
- **Multi-Language Interface**: The Cocoon's design exposes an interface that allows us to provide support for any modern day language. 
- **Scalability**: Launch multiple instances of a contract application. Incoming requests will be shared between these instances. System resource allocation can be adjusted at any time.  
- **Storage**: Contract applications are provided with the ability to store data into a ACID-complaint database service. 