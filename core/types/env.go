package types

import (
	"reflect"
	"strings"

	"github.com/ellcrys/util"
	"github.com/imdario/mergo"
)

// Env represents a collection of environment variables
type Env map[string]string

// NewEnv creates a new Env type from values in map[string]interface{}.
// and map[string]string Non-string values are omitted.
func NewEnv(val interface{}) Env {
	_env := Env{}
	switch env := val.(type) {
	case map[string]interface{}:
		for k, v := range env {
			if valStr, ok := v.(string); ok {
				_env[k] = valStr
			}
		}
	case map[string]string:
		for k, v := range env {
			_env[k] = v
		}
	}
	return _env
}

// Eql checks whether the object is equal to another
func (e Env) Eql(o Env) bool {
	return reflect.DeepEqual(e, o)
}

// ToJSON returns a json encoded representation of the ENV
func (e Env) ToJSON() []byte {
	b, _ := util.ToJSON(e)
	return b
}

// ToMap returns the object as a map type
func (e Env) ToMap() map[string]string {
	return e
}

// GetFlags returns the flags from an environment variable
func GetFlags(v string) []string {
	var varParts = strings.Split(v, "@")
	var flags = []string{}
	if len(varParts) > 1 && strings.TrimSpace(varParts[1]) != "" {
		flags = strings.Split(varParts[1], ",")
		for i, f := range flags {
			flags[i] = strings.TrimSpace(f)
		}
	}
	return flags
}

// Process applies the flags and returns both public
// environment variables and private variables.
// If keepFlag is set to true, the flags will not be removed
// in the result returned.
func (e Env) Process(keepFlags bool) (Env, Env) {
	var public = make(Env)
	var private = make(Env)

	for k, v := range e {
		flags := GetFlags(k)
		isPrivate := util.InStringSlice(flags, "private")
		for _, f := range flags {
			switch f {
			case "genRand16":
				v = util.CryptoRandKey(16)
			case "genRand24":
				v = util.CryptoRandKey(24)
			case "genRand32":
				v = util.CryptoRandKey(32)
			case "genRand64":
				v = util.CryptoRandKey(64)
			case "genRand128":
				v = util.CryptoRandKey(128)
			case "genRand256":
				v = util.CryptoRandKey(256)
			case "genRand512":
				v = util.CryptoRandKey(512)
			}
		}
		if isPrivate {
			if !keepFlags {
				private[strings.Split(k, "@")[0]] = v
			} else {
				private[k] = v
			}
		} else {
			if !keepFlags {
				public[strings.Split(k, "@")[0]] = v
			} else {
				public[k] = v
			}
		}
	}

	return public, private
}

// ProcessAsOne is like Process but returns the public and private environments merged as one Env
func (e Env) ProcessAsOne(keepFlags bool) Env {
	pub, priv := e.Process(keepFlags)
	mergo.Merge(&pub, priv)
	return pub
}
