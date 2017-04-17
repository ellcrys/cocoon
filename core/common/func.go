package common

import (
	"fmt"
	"regexp"
	"time"

	"strings"

	"net"

	"github.com/asaskevich/govalidator"
	"github.com/ncodes/cocoon/core/types"
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
	match, _ := regexp.Match("(?i)^[.a-z0-9_-]+$", []byte(name))
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

// JSONCoerceErr returns an error about an error converting json data to an object
func JSONCoerceErr(objName string, err error) error {
	return fmt.Errorf("failed to coerce response data to %s object. %s", objName, err)
}

// GetShortID gets the short version of an id.
func GetShortID(id string) string {
	if govalidator.IsUUIDv4(id) {
		return strings.Split(id, "-")[0]
	}
	if len(id) <= 12 {
		return id
	}
	if govalidator.IsEmail(id) {
		return id
	}
	return id[0:11]
}

// IsValidACLTarget checks whether an ACL target format is valid
func IsValidACLTarget(target string) error {
	if target == "*" {
		return nil
	} else if len(target) > 0 && len(strings.Split(target, ".")) < 3 {
		return nil
	}
	return fmt.Errorf("format is invalid")
}

// ResolveFirewall performs a reverse lookup of non-IP destination
// of every rule. For each resolved rule, a new cloned rule is added
// with the looked up IP used as the destination.
func ResolveFirewall(rules types.Firewall) (types.Firewall, error) {
	var newResolvedFirewall = types.Firewall{}
	for i, rule := range rules {
		if !govalidator.IsIP(rule.Destination) {
			IPs, err := net.LookupHost(rule.Destination)
			if err != nil {
				if strings.Contains(err.Error(), "no such host") {
					return nil, fmt.Errorf("rule %d: %s", i, err)
				}
				return nil, err
			}
			for _, ip := range IPs {
				if govalidator.IsIPv4(ip) {
					newResolvedFirewall = append(newResolvedFirewall, &types.FirewallRule{
						Destination:     ip,
						DestinationPort: rule.DestinationPort,
						Protocol:        rule.Protocol,
					})
					continue
				}
			}
		} else {
			newResolvedFirewall = append(newResolvedFirewall, rule)
		}
	}
	return newResolvedFirewall, nil
}

// RemoveASCIIColors takes a byte slice and remove ASCII colors
func RemoveASCIIColors(s []byte) []byte {
	return regexp.MustCompile("\033\\[[0-9;]*m").ReplaceAll(s, []byte(""))
}
