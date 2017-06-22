package archiver

import (
	"fmt"

	"github.com/ellcrys/util"
	"github.com/goware/urlx"
	cutil "github.com/ncodes/cocoon-util"
	"github.com/pkg/errors"
)

// GitObject defines a structure for fetching a git object
type GitObject struct {
	repoURL string
	version string
}

// NewGitObject creates a git object
func NewGitObject(url, version string) *GitObject {
	return &GitObject{
		repoURL: url,
		version: version,
	}
}

// getRepoTarballURL returns the repo tarball url
func (o *GitObject) getRepoTarballURL() (string, error) {

	url, err := urlx.Parse(o.repoURL)
	if err != nil {
		return "", errors.Wrap(err, "invalid repo url")
	}

	if cutil.IsGithubCommitID(o.version) {
		return fmt.Sprintf("https://api.github.com/repos%s/tarball/%s", url.Path, o.version), nil
	}

	tarballURL, err := cutil.GetGithubRepoRelease(o.repoURL, o.version)
	if err != nil {
		return "", errors.Wrap(err, "failed to get release tarball")
	}

	return tarballURL, nil
}

// Read fetches the git repository and passes the contents
// to the callback
func (o *GitObject) Read(cb func(d []byte) error) error {

	if !cutil.IsGithubRepoURL(o.repoURL) {
		return fmt.Errorf("repo url is invalid")
	}

	tarballURL, err := o.getRepoTarballURL()
	if err != nil {
		return err
	}

	if err := util.DownloadURLToFunc(tarballURL, func(chunk []byte, status int) error {
		if status == 404 {
			return fmt.Errorf("git repo tarball not found")
		}
		if err := cb(chunk); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return errors.Wrap(err, "failed to fetch tarball")
	}

	return nil
}
