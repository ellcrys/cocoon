job "connector" {
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
      driver = "raw_exec"
      
      config {
        command = "bash"
        args = ["local/runner.sh"]
      }

      artifact {
        source = "https://rawgit.com/ncodes/cocoon/master/scripts/runner.sh"
      }

      logs {
        max_files     = 10
        max_file_size = 10
      } 
      
      env {
        COCOON_ID = "abc"
        CONTAINER_ID = "abc"
        COCOON_CODE_URL = "https://github.com/ncodes/cocoon-example-01"
        COCOON_CODE_TAG = ""
        COCOON_CODE_LANG = "go"
        COCOON_BUILD_PARAMS = "eyAicGtnX21nciI6ICJnbGlkZSIgfQ=="
        COCOON_DISK_LIMIT = "1024"
        
        # The name of the connector runner script and a link to the script.
        # The runner script will fetch and run whatever is found in this environment vars.
        RUN_SCRIPT_NAME = "run-connector.sh"
        RUN_SCRIPT_URL = "https://rawgit.com/ncodes/cocoon/master/scripts/run-connector.sh"
      }

      resources {
        cpu    = 500
        memory = 1024
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
  }
}