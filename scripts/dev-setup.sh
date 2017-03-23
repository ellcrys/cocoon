# Bootstrap the entire platform for development or test

# Run the orderer
nomad run config/orderer/orderer.nomad

# Run the API
nomad run config/api/api.nomad