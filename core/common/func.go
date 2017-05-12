package common

import (
	"fmt"
	"os/signal"
	"regexp"
	"syscall"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"net"
	"strings"

	"github.com/imdario/mergo"

	"os"

	"github.com/asaskevich/govalidator"
	"github.com/ncodes/cocoon/core/lock/consul"
	"github.com/ncodes/cocoon/core/lock/memory"
	"github.com/ncodes/cocoon/core/types"
	context "golang.org/x/net/context"
)

// GetRPCErrDesc takes a grpc generated error and returns the description.
// If error is not grpc generated, it returns err.Error().
func GetRPCErrDesc(err error) string {
	return grpc.ErrorDesc(err)
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

// IsValidCocoonID checks whether a name is a accepted
// cocoon name
func IsValidCocoonID(name string) bool {
	match, _ := regexp.Match("(?i)^[a-z0-9_-]+$", []byte(name))
	return match
}

// ReRunOnError runs a function multiple times till it returns a non-error (nil) value.
// It accepts a limit on how many times to attempt a re-run and a delay duration.
// If delay is nil, there is no delay. It returns the last error of the last
// re-run attempt if the function did not succeed.
func ReRunOnError(f func() error, times int, delay time.Duration) error {
	var err error
	for i := 0; i < times; i++ {
		err = f()
		if err == nil {
			return nil
		}
		if times-1 > 1 {
			time.Sleep(delay)
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
		if rule == (types.FirewallRule{}) {
			continue
		}
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
					newResolvedFirewall = append(newResolvedFirewall, types.FirewallRule{
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

// Round rounds float values to the nearest integer
func Round(val float64) int {
	if val < 0 {
		return int(val - 0.5)
	}
	return int(val + 0.5)
}

// NewLock creates a lock. If `DEV_MEM_LOCK` is set, a
// memory lock will be returned
func NewLock(key string) (types.Lock, error) {
	if os.Getenv("DEV_MEM_LOCK") != "" {
		return memory.NewLock(key), nil
	}
	return consul.NewLock(key)
}

// NewLockWithTTL creates a lock with a custom ttl. If `DEV_MEM_LOCK` is set, a
// memory lock will be returned
func NewLockWithTTL(key string, ttl time.Duration) (types.Lock, error) {
	if os.Getenv("DEV_MEM_LOCK") != "" {
		return memory.NewLockWithTTL(key, ttl), nil
	}
	return consul.NewLockWithTTL(key, ttl)
}

// SimpleGRPCError returns simplified, user-friendly grpc errors
func SimpleGRPCError(serviceName string, err error) error {
	if err == nil {
		return nil
	}
	switch grpc.Code(err).String() {
	case "Unavailable":
		return fmt.Errorf("%s could not be reached", serviceName)
	default:
		return fmt.Errorf("rpc->%s: %s", serviceName, grpc.ErrorDesc(err))
	}
}

// ReturnFirstIfDiffInt returns the first param (a) if it
// is different from the (b) and not zero, otherwise b is returned
func ReturnFirstIfDiffInt(a, b int) int {
	if a != b && a > 0 {
		return a
	}
	return b
}

// ReturnFirstIfDiffString returns the first param (a) if it
// is different from the (b), otherwise b is returned
func ReturnFirstIfDiffString(a, b string, secondIfEmpty bool) string {
	if a != b && !(secondIfEmpty && len(a) == 0) {
		return a
	}
	return b
}

// ReturnFirstIfDiffEnv returns the first param (a) if it
// is different from the (b), otherwise b is returned
func ReturnFirstIfDiffEnv(a, b types.Env) types.Env {
	if !a.Eql(b) {
		return a
	}
	return b
}

// ReturnFirstIfDiffACLMap returns the first param (a) if it
// is different from the (b), otherwise b is returned
func ReturnFirstIfDiffACLMap(a, b types.ACLMap) types.ACLMap {
	if !a.Eql(b) {
		return a
	}
	return b
}

// GetAuthToken returns authorization code of a specific bearer from a context
func GetAuthToken(ctx context.Context, scheme string) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", fmt.Errorf("failed to get metadata from context")
	}

	authorization := md["authorization"]
	if len(authorization) == 0 {
		return "", fmt.Errorf("authorization not included in context")
	}

	authSplit := strings.SplitN(authorization[0], " ", 2)
	if len(authSplit) != 2 {
		return "", fmt.Errorf("authorization format is invalid")
	} else if authSplit[0] != scheme {
		return "", fmt.Errorf("Request unauthenticated with %s", scheme)
	}

	return authSplit[1], nil
}

// HasEnv checks whether a slice of environment variable have been set.
// It returns the variables that haven't been set
func HasEnv(envs ...string) []string {
	notSet := []string{}
	for _, v := range envs {
		if os.Getenv(v) == "" {
			notSet = append(notSet, v)
		}
	}
	return notSet
}

// OnTerminate calls a function when a terminate or interrupt signal is received.
func OnTerminate(f func(s os.Signal)) {
	sigs := make(chan os.Signal)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		s := <-sigs
		f(s)
	}()
}

// MergeMapSlice merges a slice of maps into a single map with the
// each successive maps overwriting previously available keys
func MergeMapSlice(s []map[string]interface{}) map[string]interface{} {
	var newMap map[string]interface{}
	for i := 0; i < len(s); i++ {
		mergo.MergeWithOverwrite(&newMap, s[i])
	}
	return newMap
}
