package acl

import "github.com/ncodes/cocoon/core/types"
import "fmt"

// CheckACL checks whether a cocoon is allowed to access or perform
// a specific operation against a target cocoon's resource
func CheckACL(cocoonID, targetCocoonID, opName string) error {

	// Check system ACL rules
	if targetCocoonID == types.SystemCocoonID {
		if opName == types.TxCreateLedger || opName == types.TxPut {
			return fmt.Errorf("link permission denied: CREATE/PUT operation not allowed")
		}
	}

	return nil
}
