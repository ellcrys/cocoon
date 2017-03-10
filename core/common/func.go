package common

import "strings"
import "fmt"

// StripRPCErrorPrefix takes an error return from the RPC client and removes the
// prefixed `rpc error: code = 2 desc =` statement
func StripRPCErrorPrefix(err []byte) []byte {
	return []byte(strings.TrimSpace(strings.Replace(string(err), "rpc error: code = 2 desc =", "", -1)))
}

// ToRPCError creates an RPC error message
func ToRPCError(code int, err error) error {
	return fmt.Errorf("rpc error: code = %d desc = %s", code, err)
}
