package languages

import "github.com/ncodes/modo"

// Language defines a cocoon code language
// and its unique deployment procedure.
type Language interface {
	GetName() string
	GetImage() string
	GetDownloadDestination() string
	GetCopyDestination() string
	GetSourceRootDir() string
	RequiresBuild() bool
	GetBuildScript() *modo.Do
	GetRunCommand() *modo.Do
	SetBuildParams(map[string]interface{}) error
	SetRunEnv(env map[string]string)
}
