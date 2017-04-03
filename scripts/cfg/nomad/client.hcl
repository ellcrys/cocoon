log_level = "DEBUG"

data_dir = "/tmp/client1"

client {
    enabled = true
    
    servers = ["127.0.0.1:4647"]
    
    options = {
        "driver.raw_exec.enable" = "1"
    }
}

ports {
    http = 5656
}
