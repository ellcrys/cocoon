# Bootstrap the entire platform for development or test

# Run the orderer
nomad run cfg/orderer/orderer.nomad

# Run the API
nomad run cfg/api/api.nomad