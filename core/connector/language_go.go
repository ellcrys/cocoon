package connector

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
	req         *Request
	name        string
	image       string
	userHome    string
	imgGoPath   string
	repoURL     string
	env         map[string]string
	buildParams map[string]interface{}
}

// NewGo returns a new instance a golang
// cocoon code deployment helper.
func NewGo(req *Request) *Go {
	g := &Go{
		name:      "go",
		image:     "ncodes/launch-go",
		userHome:  os.Getenv("HOME"),
		imgGoPath: "/go",
		req:       req,
		env:       map[string]string{},
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

// SetRunEnv adds environment variables to be added when the run script is constructed.
func (g *Go) SetRunEnv(env map[string]string) {
	for k, v := range env {
		g.env[k] = v
	}
}

// GetDownloadDestination returns the location to save
// the downloaded go cocoon code to.
func (g *Go) GetDownloadDestination() string {
	u, _ := urlx.Parse(g.req.URL)
	repoID := strings.Trim(u.Path, "/")
	return path.Join(g.userHome, "/ccode/sources/", g.req.ID, repoID)
}

// GetCopyDestination returns the location in
// the container where the source code will be copied to
func (g *Go) GetCopyDestination() string {
	u, _ := urlx.Parse(g.req.URL)
	repoID := strings.Trim(u.Path, "/")
	return path.Join(g.imgGoPath, "src/github.com/", strings.Split(repoID, "/")[0])
}

// GetSourceRootDir returns the root directory of the cocoon source code
// in container.
func (g *Go) GetSourceRootDir() string {
	u, _ := urlx.Parse(g.req.URL)
	repoID := strings.Trim(u.Path, "/")
	return path.Join(g.imgGoPath, "src/github.com/", repoID)
}

// RequiresBuild returns true if cocoon codes written in
// go language requires a build process.
// During development, If DEV_RUN_ROOT_BIN env is set, it will return false as
// the run command will find and find the ccode binary in the repo root.
func (g *Go) RequiresBuild() bool {
	if len(os.Getenv("DEV_RUN_ROOT_BIN")) > 0 {
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
// cocoon code according to the build and installation process
// of the language. If DEV_RUN_ROOT_BIN env is set, it will run the
// ccode binary located in the mount destination.
func (g *Go) GetRunScript() []string {

	// prepare environment variables
	env := []string{}
	for k, v := range g.env {
		env = append(env, fmt.Sprintf("export %s=%s", k, v))
	}

	script := []string{
		"bash",
		"-c",
		strings.Join(env, " && ") + " && ccode",
	}

	// run ccode in the repo root if DEV_RUN_ROOT_BIN is set (for development only)
	if len(os.Getenv("DEV_RUN_ROOT_BIN")) > 0 {
		script = []string{"bash", "-c", fmt.Sprintf("cd %s && ./ccode", g.GetSourceRootDir())}
	}

	return script
}
