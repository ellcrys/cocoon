# Bootstrap the entire platform for development or test

# Run the orderer
nomad run dev_config/orderer/orderer.nomad

# Run the API
nomad run dev_config/api/api.nomad