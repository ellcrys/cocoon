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
        args = ["local/runner.sh"]
      }
      
      env {
        COCOON_ID = "abc"
        COCOON_CODE_URL = "https://github.com/ncodes/cocoon-example-01"
        COCOON_CODE_TAG = ""
        COCOON_CODE_LANG = "go"
        COCOON_BUILD_PARAMS = "eyAicGtnX21nciI6ICJnbGlkZSIgfQ=="
        COCOON_CONTAINER_NAME = "${NOMAD_TASK_NAME}-${NOMAD_ALLOC_ID}"
        
        # The name of the connector runner script and a link to the script.
        # The runner script will fetch and run whatever is found in this environment vars.
        RUN_SCRIPT_NAME = "run-connector.sh"
        RUN_SCRIPT_URL = "https://rawgit.com/ncodes/cocoon/master/scripts/run-connector.sh"
      }

      artifact {
        source = "https://rawgit.com/ncodes/cocoon/master/scripts/runner.sh"
      }
      
      logs {
        max_files     = 10
        max_file_size = 10
      }

      resources {
        cpu    = 500
        memory = 1024
        network {
          mbits = 1
          port "RPC" {}
        }
      }

      service {
        name = "connectors"
        tags = []
        port = "RPC"
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
        args = ["-c", "echo 'Hello Human. I am alive'; tail -f /dev/null"]
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
          port "RPC" {}
        }
      }
    }
  }
}