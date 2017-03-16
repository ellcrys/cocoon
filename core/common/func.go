package common

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

// StripRPCErrorPrefix takes an error return from the RPC client and removes the
// prefixed `rpc error: code = 2 desc =` statement
func StripRPCErrorPrefix(err []byte) []byte {
	return []byte(strings.TrimSpace(strings.Replace(string(err), "rpc error: code = 2 desc =", "", -1)))
}

// ToRPCError creates an RPC error message
func ToRPCError(code int, err error) error {
	return fmt.Errorf("rpc error: code = %d desc = %s", code, err)
}

// IsUniqueConstraintError checks whether an error is a postgres
// contraint error affecting a column.
func IsUniqueConstraintError(err error, column string) bool {
	if m, _ := regexp.Match(`^.*unique constraint "idx_name_`+column+`"$`, []byte(err.Error())); m {
		return true
	}
	return false
}

// IsValidResName checks whether a name is a accepted
// resource name format.
func IsValidResName(name string) bool {
	match, _ := regexp.Match("(?i)^[a-z_]+$", []byte(name))
	return match
}

// ReRunOnError runs a function multiple times till it returns a non-error (nil) value.
// It accepts a limit on how many times to attempt a re-run and a delay duration.
// If delay is nil, there is no delay. It returns the last error of the last
// re-run attempt if the function did not succeed.
func ReRunOnError(f func() error, times int, delay *time.Duration) error {
	var err error
	for i := 0; i < times; i++ {
		err = f()
		if err == nil {
			return nil
		}
		if times-1 > 1 && delay != nil {
			time.Sleep(*delay)
		}
	}
	return err
}
