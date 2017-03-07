package launcher

// Language defines a cocoon code language
// and its unique deployment procudures.
type Language interface {
	GetName() string
	GetImage() string
	GetDownloadDestination() string
	GetCopyDestination() string
	GetSourceRootDir() string
	RequiresBuild() bool
	GetBuildScript() string
	GetRunScript() []string
	SetBuildParams(map[string]interface{}) error
}
