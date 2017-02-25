package launcher

import (
	"os"
	"os/user"
	"path"
	"strings"

	"fmt"

	"github.com/ellcrys/util"
	"github.com/goware/urlx"
)

// Go defines a deployment helper for go cocoon.
type Go struct {
	name        string
	image       string
	userHome    string
	imgGoPath   string
	repoURL     string
	buildParams map[string]interface{}
}

// NewGo returns a new instance a golang
// cocoon code deployment helper.
func NewGo() *Go {
	g := &Go{
		name:      "go",
		image:     "ncodes/launch-go",
		userHome:  os.Getenv("HOME"),
		imgGoPath: "/go",
	}
	g.setUserHomeDir()
	return g
}

// GetName returns the name for the language
func (g *Go) GetName() string {
	return g.name
}

// GetImage returns the image name to deploy cocoon code on
func (g *Go) GetImage() string {
	return g.image
}

// setUserHomeDir sets the user home directory
func (g *Go) setUserHomeDir() error {
	usr, err := user.Current()
	if err != nil {
		return err
	}
	g.userHome = usr.HomeDir
	return nil
}

// GetDownloadDestination returns the location to save
// the downloaded go cocoon code to.
func (g *Go) GetDownloadDestination(url string) string {
	g.repoURL = url
	u, _ := urlx.Parse(url)
	repoID := strings.Trim(u.Path, "/")
	return path.Join(g.userHome, "/ccode/source", repoID)
}

// GetMountDestination returns the location in
// the container where the source code root will be mounted
func (g *Go) GetMountDestination(url string) string {
	u, _ := urlx.Parse(url)
	repoID := strings.Trim(u.Path, "/")
	return path.Join(g.imgGoPath, "src/github.com/", repoID)
}

// RequiresBuild returns true if cocoon codes written in
// go langugage requires a build process.
// During development, If RUN_ROOT_BIN env is set, it will return false as
// the run command will find and find the ccode binary in the repo root.
func (g *Go) RequiresBuild() bool {
	if len(os.Getenv("RUN_ROOT_BIN")) > 0 {
		return false
	}
	return true
}

// SetBuildParams sets and validates build parameters
func (g *Go) SetBuildParams(buildParams map[string]interface{}) error {
	g.buildParams = buildParams
	if pkgMgr := g.buildParams["pkg_mgr"]; pkgMgr != nil {
		if !util.InStringSlice([]string{"glide"}, pkgMgr.(string)) {
			return fmt.Errorf("invalid pkg_mgr value in build script")
		}
	}
	return nil
}

// GetBuildScript will return the script required
// to create an executable
func (g *Go) GetBuildScript() string {

	// run known package manager fetch routine
	pkgFetchCmd := ""
	if pkgMgr := g.buildParams["pkg_mgr"]; pkgMgr != nil {
		switch pkgMgr {
		case "glide":
			pkgFetchCmd = "glide install"
		}
	}

	return strings.Join(util.RemoveEmptyInStringSlice([]string{pkgFetchCmd, "go build -v -o /bin/ccode"}), " && ")
}

// GetRunScript returns the script required to start the
// cocoon code accodeording to the build and installation process
// of the language. If RUN_ROOT_BIN env is set, it will run the
// ccode binary located in the mount destination.
func (g *Go) GetRunScript() []string {
	script := []string{strings.Join([]string{"ccode"}, " && ")}

	// run ccode in the repo root if RUN_ROOT_BIN is set (for development only)
	if len(os.Getenv("RUN_ROOT_BIN")) > 0 {
		script = []string{"bash", "-c", fmt.Sprintf("cd %s && ./ccode", g.GetMountDestination(g.repoURL))}
	}

	return script
}
