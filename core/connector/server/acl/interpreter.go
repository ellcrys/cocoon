package acl

import (
	"fmt"

	"strings"

	"github.com/ellcrys/util"
	"github.com/ncodes/cocoon/core/types"
)

var (
	// PrivAllow allows all operations
	PrivAllow = "allow"
	// PrivAllowCreateLedger allows a ledger creation operation
	PrivAllowCreateLedger = "allow-create-ledger"
	// PrivAllowPut allows a put operation
	PrivAllowPut = "allow-put"
	// PrivAllowGet allows a get operation
	PrivAllowGet = "allow-get"
	// PrivDeny denies all operations
	PrivDeny = "deny"
	// PrivDenyCreateLedger denies a ledger creation operation
	PrivDenyCreateLedger = "deny-create-ledger"
	// PrivDenyPut denies a put operation
	PrivDenyPut = "deny-put"
	// PrivDenyGet denies a get operation
	PrivDenyGet = "deny-get"

	// validPrivileges refers to the accepted/known privileges
	validPrivileges = []string{
		PrivAllow,
		PrivAllowCreateLedger,
		PrivAllowPut,
		PrivAllowGet,
		PrivDeny,
		PrivDenyCreateLedger,
		PrivDenyPut,
		PrivDenyGet,
	}
)

// Interpreter represents an ACL rule interpreter.
type Interpreter struct {
	rules         map[string]interface{}
	defaultPolicy bool
}

// NewInterpreter creates a new ACLInterpreter object
func NewInterpreter(rules map[string]interface{}, defaultPolicy bool) *Interpreter {
	return &Interpreter{rules: rules, defaultPolicy: defaultPolicy}
}

// NewInterpreterFromACLMap creates a new ACLInterpreter using an ACLMap
func NewInterpreterFromACLMap(rules types.ACLMap, defaultPolicy bool) *Interpreter {
	return &Interpreter{rules: rules, defaultPolicy: defaultPolicy}
}

// isValidLedgerKeyType checks whether the value type of a ledger rule is
// the expected type of string, map of string or map of interface{}
func (i *Interpreter) isValidLedgerKeyType(v interface{}) bool {
	switch v.(type) {
	case string, map[string]string, map[string]interface{}:
		return true
	default:
		return false
	}
}

// IsValidPrivilege checks whether a privilege is valid
func IsValidPrivilege(p string) bool {
	return util.InStringSlice(validPrivileges, strings.TrimSpace(strings.ToLower(p)))
}

// isValidActorID checks whether a user/cocoon id is valid
func (i *Interpreter) isValidActorID(u string) error {
	if len(u) == 0 {
		return fmt.Errorf("cocoon or identity id cannot be an empty string")
	}
	return nil
}

// Validate takes acl rules and checks whether it is value
func (i *Interpreter) Validate() []error {
	var errs []error
	for ledgerName, val := range i.rules {

		if ledgerName == "*" {
			if _, ok := val.(string); !ok {
				errs = append(errs, fmt.Errorf("%s: invalid wildcard ledger value type. Expects string value", ledgerName))
			}
		}

		if !i.isValidLedgerKeyType(val) {
			errs = append(errs, fmt.Errorf("%s: invalid ledger value type. Expects string or map of strings", ledgerName))
		}

		if ruleVal, ok := val.(string); ok {
			privileges := strings.Split(ruleVal, " ")
			for _, priv := range privileges {
				if !IsValidPrivilege(priv) {
					errs = append(errs, fmt.Errorf("%s: ledger contains an invalid privilege (%s)", ledgerName, priv))
				}
			}
		}

		if actorsPrivileges, ok := val.(map[string]string); ok {
			for user, _privileges := range actorsPrivileges {
				if err := i.isValidActorID(user); err != nil {
					errs = append(errs, fmt.Errorf("%s: invalid actor id. %s", ledgerName, err))
				}
				privileges := strings.Split(_privileges, " ")
				for _, priv := range privileges {
					if !IsValidPrivilege(priv) {
						errs = append(errs, fmt.Errorf("%s: ledger actor contains an invalid privilege (%s)", ledgerName, priv))
					}
				}
			}
		}
	}
	return errs
}

// isPermitted takes a privilege and an operation and returns
// 0 if the operation is not allowed, 1 if the operation is allowed
// and -1 if the privilege and operation combo has no permission rule logic.
func (i *Interpreter) isPermitted(privilege, operation string) int {
	if privilege == PrivAllow {
		return 1
	} else if privilege == PrivDeny {
		return 0
	}
	if privilege == PrivAllowCreateLedger && operation == types.TxCreateLedger {
		return 1
	} else if privilege == PrivDenyCreateLedger && operation == types.TxCreateLedger {
		return 0
	}
	if privilege == PrivAllowGet && util.InStringSlice([]string{types.TxGet, types.TxGetByID, types.TxGetLedger, types.TxGetBlockByID, types.TxRangeGet}, operation) {
		return 1
	} else if privilege == PrivDenyGet && util.InStringSlice([]string{types.TxGet, types.TxGetByID, types.TxGetLedger, types.TxGetBlockByID, types.TxRangeGet}, operation) {
		return 0
	}
	if privilege == PrivAllowPut && operation == types.TxPut {
		return 1
	} else if privilege == PrivDenyPut && operation == types.TxPut {
		return 0
	}
	return -1
}

// IsAllowed checks whether an operation is permitted. The supported
// operations are all Tx operations in the types/transactions.go file.
// If actorID is set, the specific actor rule takes precedence over the
// wildcard rule (if set). If no rule is found for an operation and no
// wildcard ledger rule, the operation is considered allowed.
func (i *Interpreter) IsAllowed(ledgerName, actorID, operation string) bool {

	var allowed = i.defaultPolicy

	// check if wildcard ledger is set, if yes, check if it permits the operation
	if allLedgersPrivileges, ok := i.rules["*"]; ok {
		privileges := strings.Split(allLedgersPrivileges.(string), " ")
		for _, p := range privileges {
			switch i.isPermitted(strings.TrimSpace(p), operation) {
			case 0:
				allowed = false
			case 1:
				allowed = true
			}
		}
	}

	// check with ledger-specific privilege if provided
	if ledgerPrivileges, ok := i.rules[ledgerName]; ok {
		if allPrivileges, ok := ledgerPrivileges.(string); ok {
			privileges := strings.Split(allPrivileges, " ")
			for _, p := range privileges {
				switch i.isPermitted(strings.TrimSpace(p), operation) {
				case 0:
					allowed = false
				case 1:
					allowed = true
				}
			}
		}

		// check ledger-specific, actor-specific privilege if provided
		switch actorsPrivileges := ledgerPrivileges.(type) {
		case map[string]string:
			if actorPrivileges, ok := actorsPrivileges[actorID]; ok {
				privileges := strings.Split(actorPrivileges, " ")
				for _, p := range privileges {
					switch i.isPermitted(strings.TrimSpace(p), operation) {
					case 0:
						allowed = false
					case 1:
						allowed = true
					}
				}
			}
		case map[string]interface{}:
			if actorPrivileges, ok := actorsPrivileges[actorID]; ok {
				privileges := strings.Split(actorPrivileges.(string), " ")
				for _, p := range privileges {
					switch i.isPermitted(strings.TrimSpace(p), operation) {
					case 0:
						allowed = false
					case 1:
						allowed = true
					}
				}
			}
		}
	}

	return allowed
}
