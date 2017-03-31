# Increase log verbosity
log_level = "DEBUG"

# Setup data dir
data_dir = "/tmp/client1"

# Enable the client
client {
    enabled = true
    servers = ["127.0.0.1:4647"]
    chroot_env {
        "/bin" = "/bin"
        "/etc" = "/etc"
        "/lib" = "/lib"
        "/lib32" = "/lib32"
        "/lib64" = /lib64"
        "/run/resolvconf" = "/run/resolvconf"
        "/sbin" = "/sbin"
        "/usr" = "/usr"
        "/go" = "/go"
    }
}

# Modify our port to avoid a collision with server1
ports {
    http = 5656
}
