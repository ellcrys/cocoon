# A contract file describes a contract specification
contracts {
    
    # A unique ID (ex: com.mywebsite.com.myname).
    # If not provide, a UUID v4 ID is generated. 
    id = "u1"
    
    # Contract source location and information
    repo {
        # The pubic github repository
        url = "https://github.com/ncodes/cocoon-example-01" 
        # The github release tag or commit id (default: latest release)
        version = "f66421c591e0675ee721108f64c118399213447d"
        # The contract source code language
        language = "go"
        # Specify the ID of another cocoon to link to.
        # The contract will have the same privileges of the linked contract.
        # Both contracts must be owned by same identity
        link = ""
    }
    
    # Provide build information if the contract code requires it
    build {
        # The package manager to use (supported: glide, govendor)
        pkgMgr = "govendor"
    }
    
    # Resources to allocate to the contract's cocoon
    resources {
       resource_set = "s1"
    }
    
    # Provide signatory information
    signatories {
        # The maximum number of signatories to accept
        max = 1
        # The number of signature required to approve a release
        threshold = 1
    }
    
    # Access control list stanza allows the contract
    # to allow or deny access to perform specific operations by other contracts.
    acl {
        # Allow all operations but deny the ability to create ledgers
        "*" = "allow deny-create-ledger"
    }
    
    # Firewall stanza determines the addresses the contract
    # can make outbound connections to.
    firewall {
        destination = "google.com"
        destinationPort = "80"
        protocol = "tcp"
    }
}
