job "orderer" {
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

  group "orderers" {
    count = 1

    restart {
      attempts = 10
      interval = "30s"
      delay = "25s"
      mode = "delay"
    }

    ephemeral_disk {
      size = 300
    }

    task "redis" {
      driver = "docker"
      config {
        image = "ncodes/cocoon-launcher:latest"
        port_map {}
      }

    //   artifact {
    //     source = "http://foo.com/artifact.tar.gz"
    //     options {
    //       checksum = "md5:c4aa853ad2215426eb7d70a21922e794"
    //     }
    //   }

      logs {
        max_files     = 10
        max_file_size = 10
      }

      resources {
        cpu    = 500
        memory = 256 
        network {
          mbits = 1000
          port "orderer-grpc" {}
        }
      }

      service {
        name = "orderer"
        tags = []
        port = "orderer-grpc"
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