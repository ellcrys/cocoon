package common

import (
	"fmt"
	"regexp"
	"time"

	"strings"
)

// GetRPCErrDesc takes a grpc generated error and returns the description.
// If error is not grpc generated, it returns err.Error().
func GetRPCErrDesc(err error) string {
	if strings.HasPrefix(strings.ToLower(err.Error()), "rpc error:") {
		ss := strings.Split(err.Error(), "=")
		return strings.TrimSpace(ss[len(ss)-1])
	}
	return err.Error()
}

// CompareErr compares two error string values. Returns 0 if equal.
// Removes GRPC prefixes if available.
func CompareErr(errA, errB error) int {
	return strings.Compare(GetRPCErrDesc(errA), GetRPCErrDesc(errB))
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
	match, _ := regexp.Match("(?i)^[a-z0-9_]+$", []byte(name))
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

// MBToByte returns the amount of bytes in a MB.
func MBToByte(mb int64) int64 {
	return (1048576 * mb)
}

// CapitalizeString capitalizes the first character of all sentences in a given string.
func CapitalizeString(str string) string {
	var re = regexp.MustCompile(`^([a-z]{1})|\.[ ]+?([a-z]{1})`)
	return re.ReplaceAllStringFunc(str, func(c string) string {
		return strings.ToUpper(c)
	})
}
