# Bootstrap the entire platform for development or test

consul agent -dev & sudo nomad agent -config=config/nomad/server.hcl & sudo nomad agent -config=config/nomad/client.hcl

# Run the orderer
nomad run config/orderer/orderer.nomad

# Run the API
nomad run config/api/api.nomad