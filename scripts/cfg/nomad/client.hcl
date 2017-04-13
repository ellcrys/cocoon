log_level = "DEBUG"

data_dir = "/tmp/client1"

client {
    enabled = true
    
    servers = ["127.0.0.1:4647"]
    
    options = {
        "docker.privileged.enabled" = "true"
    }
}

ports {
    http = 5656
}
