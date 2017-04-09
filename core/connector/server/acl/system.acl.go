package acl

import "fmt"

// SystemACL represents the system's access control list.
// Here we are only disabling the ability to create a ledger
// and to put a transaction on all of the system's ledgers.
var SystemACL = map[string]interface{}{
	"*": fmt.Sprintf("%s %s", PrivDenyCreateLedger, PrivDenyPut),
}
