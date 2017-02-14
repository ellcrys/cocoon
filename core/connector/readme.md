### Connector
The connector is a cocoon agent tasked with running the cocoon code (a.k.a smart contract or chaincode) in a
docker container within the cluster. It is responsible for starting and managing the cocoon code and in turn 
provides access to services to the cocoon code via GRPC (and/or REST?). 

## Supported Languages
These are supported languages.

- Go
- Node.js (Coming soon)
- Ruby (Coming soon)

## Required Environment Variable
*COCOON_ID* - The unique id for the cocoon. 
*COCOON_CODE_URL* - The github repo url of the cocoon code.
*COCOON_CODE_TAG* - The github repo release to install
*COCOON_CODE_LANG* - The github repo language (e.g go).