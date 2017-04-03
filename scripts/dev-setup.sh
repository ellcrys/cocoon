# Bootstrap the entire platform for development or test

# Start consul and nomad server, client
nohup 'consul agent -dev & sudo nomad agent -config=cfg/nomad/server.hcl & sudo nomad agent -config=cfg/nomad/client.hcl &'

# Run the orderer
nomad run cfg/orderer/orderer.nomad

# Run the API
nomad run cfg/api/api.nomad