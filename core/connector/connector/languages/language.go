package languages

import "github.com/ncodes/modo"

// Language defines a cocoon code language
// and its unique deployment procedure.
type Language interface {

	// GetName returns the name of the language
	GetName() string

	// GetImage returns the docker image suitable for running the language
	GetImage() string

	// GetDownloadDir returns the directory path to store downloaded source code
	GetDownloadDir() string

	// GetCopyDir returns the path to copy the downloaded source from the
	// path returned by GetDownloadDir()
	GetCopyDir() string

	// GetSourceDir returns the path to the root of the downloaded source code
	GetSourceDir() string

	// RequiresBuild returns true if the language requires a build stage
	RequiresBuild() bool

	// GetBuildCommand returns the build command to run
	GetBuildCommand() *modo.Do

	// GetRunCommand returns the program execution command
	GetRunCommand() *modo.Do

	// SetBuildParams sets the build parameters required to construct the build command
	// returned by GetBuildCommand()
	SetBuildParams(map[string]interface{}) error

	// SetRunEnv sets the environment to apply when constructing the run/execution command
	// returned by GetRunCommand()
	SetRunEnv(env map[string]string)
}
