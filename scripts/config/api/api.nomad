job "api" {
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

  group "apis" {
    count = 1

    restart {
      attempts = 5
      interval = "30s"
      delay = "5s"
      mode = "delay"
    }

    ephemeral_disk {
      size = 300
    }

    task "api" {
      driver = "docker"
      
      config {
        image = "ncodes/cocoon-launcher:latest"
        command = "bash"
        args = ["run.sh"]
        work_dir = "/local/scripts"
        network_mode = "host"
        port_map {
            API_RPC = 8005
        }
      }

      artifact {
        source = "https://raw.githubusercontent.com/ncodes/cocoon/master/scripts/config/api/run.sh"
        destination = "/local/scripts"
      }

      logs {
        max_files     = 10
        max_file_size = 10
      }
      
      env {
          CONSUL_ADDR = "localhost:8500"
          API_SIGN_KEY = "x/A%D*G-KaPdSgVkYp3s6v9y$B&E(H+MbQeThWmZq4t7w!z%C*F-J@NcRfUjXn2r",
      }

      resources {
        cpu    = 500
        memory = 256 
        network {
          mbits = 100
          port "API_RPC" {}
        }
      }

      service {
        name = "apis"
        tags = []
        port = "API_RPC"
        check {
          name     = "alive"
          type     = "tcp"
          interval = "10s"
          timeout  = "2s"
        }
      }
    }
  }
}