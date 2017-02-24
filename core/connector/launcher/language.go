package launcher

// Language defines a cocoon code language
// and its unique deployment procudures.
type Language interface {
	GetName() string
	GetImage() string
	GetDownloadDestination(string) string
	GetMountDestination(string) string
	RequiresBuild() bool
	GetBuildScript(map[string]interface{}) string
	GetRunScript() []string
}
