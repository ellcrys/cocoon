package launcher

import (
	"fmt"
	"os"
	"os/exec"

	"path"

	"github.com/ncodes/cocoon/core/others"
	logging "github.com/op/go-logging"
)

var log = logging.MustGetLogger("launcher")

// Launcher defines cocoon code
// deployment service.
type Launcher struct {
	failed    chan bool
	languages []Language
}

// NewLauncher creates a new launcher
func NewLauncher(failed chan bool) *Launcher {
	return &Launcher{
		failed: failed,
	}
}

// failed sends true or false to the
// launch failed channel to indicate success or failure
func (lc *Launcher) setFailed(v bool) {
	lc.failed <- v
}

// Launch starts a cocoon code
func (lc *Launcher) Launch(req *Request) {

	log.Info("Ready to install cocoon code")
	log.Debugf("Found ccode url=%s and lang=%s", req.URL, req.Lang)

	lang := lc.GetLanguage(req.Lang)
	if lang == nil {
		log.Errorf("cocoon code language (%s) not supported", req.Lang)
		lc.setFailed(true)
		return
	}

	_, err := lc.fetchSource(req, lang)
	if err != nil {
		log.Error(err.Error())
		lc.setFailed(true)
		return
	}

	log.Info(lang)
}

// AddLanguage adds a new langauge to the launcher.
// Will return error if language is already added
func (lc *Launcher) AddLanguage(lang Language) error {
	if lc.GetLanguage(lang.GetName()) != nil {
		return fmt.Errorf("language already exist")
	}
	lc.languages = append(lc.languages, lang)
	return nil
}

// GetLanguage will return a langauges or nil if not found
func (lc *Launcher) GetLanguage(name string) Language {
	for _, l := range lc.languages {
		if l.GetName() == name {
			return l
		}
	}
	return nil
}

// GetLanguages returns all languages added to the launcher
func (lc *Launcher) GetLanguages() []Language {
	return lc.languages
}

// fetchSource fetches the cocoon code source from
// a remote address
func (lc *Launcher) fetchSource(req *Request, lang Language) (string, error) {

	if !others.IsGithubRepoURL(req.URL) {
		return "", fmt.Errorf("only public source code hosted on github is supported") // TODO: support zip files
	}

	lc.fetchFromGit(req, lang)

	return "", nil
}

// findLaunch looks for a previous stored launch/Redeployment by id
// TODO: needs implementation
func (lc *Launcher) findLaunch(id string) interface{} {
	return nil
}

// fetchFromGit fetchs cocoon code from git repo.
// and returns the download directory.
func (lc *Launcher) fetchFromGit(req *Request, lang Language) (string, error) {

	var repoTarURL, downloadDst, unpackDst string
	var err error

	// checks if job was previously deployed. find a job by the job name.
	if lc.findLaunch(req.ID) != nil {
		return "", fmt.Errorf("cocoon code was previously launched") // TODO: fetch last launch tag and use it
	}

	repoTarURL, err = others.GetGithubRepoRelease(req.URL, req.Tag)
	if err != nil {
		return "", fmt.Errorf("Failed to fetch release from github repo. %s", err)
	}

	// set tag to latest if not provided
	tagStr := req.Tag
	if tagStr == "" {
		tagStr = "latest"
	}

	// determine download directory
	downloadDst = lang.GetDownloadDestination(req.URL)
	unpackDst = downloadDst

	// delete download directory if it exists
	if _, err := os.Stat(downloadDst); err == nil {
		// log.Info("Download destination is not empty. Deleting content")
		// if err = os.RemoveAll(downloadDst); err != nil {
		// 	return "", fmt.Errorf("failed to delete contents of download directory")
		// }
		// log.Info("Download directory has been deleted")
	}

	log.Info("Downloading cocoon repository with tag=%s, dst=%s", tagStr, downloadDst)
	filePath := path.Join(downloadDst, fmt.Sprintf("%s.tar.gz", req.ID))
	err = others.DownloadFile(repoTarURL, filePath, func(buf []byte) {})
	if err != nil {
		return "", err
	}

	log.Info("Successfully downloaded cocoon code")
	log.Debugf("Unpacking cocoon code to %s", filePath)

	if err = os.MkdirAll(unpackDst, os.ModePerm); err != nil {
		return "", fmt.Errorf("Failed to create unpack directory. %s", err)
	}

	// unpack tarball
	cmd := "tar"
	args := []string{"-xf", filePath, "-C", unpackDst, "--strip-components", "1"}
	if err = exec.Command(cmd, args...).Run(); err != nil {
		return "", fmt.Errorf("Failed to unpack cocoon code repo tarball. %s", err)
	}

	log.Infof("Successfully unpacked cocoon code to %s", unpackDst)

	os.Remove(filePath)
	log.Info("Deleted the cocoon code tarball")

	return unpackDst, nil
}
