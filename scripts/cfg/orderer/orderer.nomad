job "orderer" {
  datacenters = ["dc1"]
  region = "global"
  type = "service"
  constraint {
    #attribute = "${attr.kernel.name}"
    #value     = "linux"
  }
  update {
    stagger = "10s"
    max_parallel = 1
  }

  group "orderers" {
    count = 1

    restart {
      attempts = 5
      interval = "30s"
      delay = "5s"
      mode = "delay"
    }

    ephemeral_disk {
      size = 1024
    }
    
     meta {
        VERSION = "master"
    }

    task "orderer" {
      driver = "docker"
      
      config {
        image = "ncodes/cocoon-launcher:latest"
        command = "bash"
        args = ["run.sh"]
        work_dir = "/local/scripts"
        network_mode = "host"
        port_map {}
      }

      artifact {
        source = "https://raw.githubusercontent.com/ncodes/cocoon/${NOMAD_META_VERSION}/scripts/cfg/orderer/run.sh"
        destination = "/local/scripts"
      }

      logs {
        max_files     = 10
        max_file_size = 10
      }
      
      env {
        ENV = "production"
        ORDERER_VERSION = "1.0.0"   
        STORE_CON_STR = "host=localhost user=postgres dbname=cocoon sslmode=disable password="
      }

      resources {
        cpu    = 500
        memory = 1024
        network {
          mbits = 1
          port "ORDERER_RPC" {}
        }
      }

      service {
        name = "orderers"
        tags = []
        port = "ORDERER_RPC"
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