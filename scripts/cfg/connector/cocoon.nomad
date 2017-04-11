job "cocoon" {
  datacenters = ["dc1"]
  region = "global"
  type = "service"
  
  constraint {
    attribute = "${attr.kernel.name}"
    value     = "linux"
  }
  
  update {
    stagger = "10s"
    max_parallel = 1
  }

  group "cocoon-grp" {
    count = 1

    restart {
      attempts = 5
      interval = "30s"
      delay = "5s"
      mode = "delay"
    }
 
    task "connector" {
      driver = "docker"
      
      config {
        network_mode = "host"
        privileged = true
        force_pull = true
        image = "ncodes/cocoon-launcher:latest"  
        command = "bash"
        args = ["-c", "echo 'Sleeping'; sleep 3600"]
      }
      
      env {
        COCOON_ID = "abc"
        COCOON_CODE_URL = "https://github.com/ncodes/cocoon-example-01"
        COCOON_CODE_TAG = ""
        COCOON_CODE_LANG = "go"
        COCOON_BUILD_PARAMS = "eyAicGtnX21nciI6ICJnbGlkZSIgfQ=="
      }

      logs {
        max_files     = 10
        max_file_size = 10
      }

      resources {
        cpu    = 500
        memory = 512
        network {
          mbits = 1
          port "CONNECTOR_RPC" {}
          port "COCOON_RPC" {}
        }
      }

      service {
        name = "connectors"
        tags = []
        port = "CONNECTOR_RPC"
        check {
          name     = "alive"
          type     = "tcp"
          interval = "10s"
          timeout  = "2s"
        }
      }
    
    
    }
    
    task "code" {
      driver = "docker"
      
      config {
        network_mode = "bridge"
        force_pull = true
        image = "ncodes/launch-go:latest"  
        command = "bash"
        args = ["-c", "echo 'Sleeping'; sleep 3600"]
      }

      logs {
        max_files     = 10
        max_file_size = 10
      }

      resources {
        cpu    = 500
        memory = 512
        network {
          mbits = 1
        }
      }
    }
  }
}