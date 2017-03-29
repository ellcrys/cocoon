job "connector" {
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

  group "connectors" {
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

    task "connector" {
      driver = "docker"
      
      config {
        image = "ncodes/cocoon-launcher:latest"
        command = "bash"
        args = ["${NOMAD_META_SCRIPTS_DIR}/${NOMAD_META_DEPLOY_SCRIPT_NAME}"]
        work_dir = "/local/scripts"
        network_mode = "host"
        privileged = true
        port_map {}
        volumes = [
            "/tmp:/tmp2",
        ]
      }

      artifact {
        source = "https://raw.githubusercontent.com/ncodes/cocoon/connector-redesign/scripts/${NOMAD_META_DEPLOY_SCRIPT_NAME}"
        destination = "/local/scripts"
      }

      logs {
        max_files     = 10
        max_file_size = 10
      }
      
      env {
        COCOON_ID = "abc"
        COCOON_CODE_URL = "https://github.com/ncodes/cocoon-example-01"
        COCOON_CODE_TAG = ""
        COCOON_CODE_LANG = "go"
        COCOON_BUILD_PARAMS = "eyAicGtnX21nciI6ICJnbGlkZSIgfQ=="
        COCOON_DISK_LIMIT = "1024"
        GLIDE_TMP = "/tmp2"
      }

      resources {
        cpu    = 500
        memory = 1024
        network {
          mbits = 1
          port "CONNECTOR_RPC" {}
        }
      }
      
      meta {
          DEPLOY_SCRIPT_NAME = "run-connector.sh",
          SCRIPTS_DIR = "/local/scripts",
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
  }
}