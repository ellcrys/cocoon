log_level = "DEBUG"

data_dir = "/tmp/client1"

client {
    enabled = true
    
    servers = ["127.0.0.1:4647"]
    
    chroot_env {
        "/bin" = "/bin"
        "/etc" = "/etc"
        "/lib" = "/lib"
        "/lib32" = "/lib32"
        "/lib64" = "/lib64"
        "/run/resolvconf" = "/run/resolvconf"
        "/sbin" = "/sbin"
        "/usr" = "/usr"
    }
}

ports {
    http = 5656
}
