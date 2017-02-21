package launcher

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/goware/urlx"
	"github.com/ncodes/cocoon/core/others"
	logging "github.com/op/go-logging"
)

var log = logging.MustGetLogger("launcher")
var cmdCCode *exec.Cmd

// GetJob returns the job details if
// it was previously deployed
func GetJob(jobName string) string {
	// not implemented
	return ""
}

// Launch downloads, build and starts a cocoon code.
// It looks at the cocoon code parameters specified
// in the environement required to launch the cocoon code.
// It passes true to the failed channel if launch failed,
// otherwise true
func Launch(failed chan bool) {

	log.Info("Ready to install cocoon code")

	// get cocoon code github link and language
	ccID := os.Getenv("COCOON_ID")
	ccURL := os.Getenv("COCOON_CODE_URL")
	ccTag := os.Getenv("COCOON_CODE_TAG")
	ccLang := os.Getenv("COCOON_CODE_LANG")

	if ccID == "" {
		log.Error("Cocoon code id not set @ $COCOON_ID")
		failed <- true
		return
	} else if ccURL == "" {
		log.Error("Cocoon code url not set @ $COCOON_CODE_URL")
		failed <- true
		return
	} else if ccLang == "" {
		log.Error("Cocoon code url not set @ $COCOON_CODE_LANG")
		failed <- true
		return
	}

	log.Debugf("Found ccode url=%s and lang=%s", ccURL, ccLang)

	if others.IsGithubRepoURL(ccURL) {

		// download release from github
		repoDir, err := fetchFromGit(ccURL, ccTag, ccLang)
		if err != nil {
			log.Error(err.Error())
			failed <- true
			return
		}

		// start cocoon code
		start(ccLang, repoDir)

		failed <- false
		return
	}

	log.Errorf("Cocoon url [%s] is not a github repo link", ccURL)
	failed <- true
}

// fetchFromGit fetchs cocoon code from git repo.
// and returns the download directory.
func fetchFromGit(url, tag, lang string) (string, error) {

	var prevJob, repoTarURL, downloadDst, unpackDst string
	var err error

	// checks if job was previously deployed. find a job by the job name.
	prevJob = GetJob(os.Getenv("COCOON_ID"))
	if prevJob == "" {
		repoTarURL, err = others.GetGithubRepoRelease(url, tag)
		if err != nil {
			return "", fmt.Errorf("Failed to fetch release from github repo. %s", err)
		}
	} else {
		// repoID = prevRepoID
		return "", fmt.Errorf("Job previously deployed. Redeployment not implemented yet")
	}

	// set tag to latest if not provided
	tagStr := tag
	if tag == "" {
		tagStr = "latest"
	}

	// determine download directory
	u, _ := urlx.Parse(url)
	username := strings.Split(strings.Trim(u.Path, "/"), "/")[0]
	if lang == "go" {
		gopath := os.Getenv("GOPATH")
		downloadDst = fmt.Sprintf("%s/src/github.com/%s", gopath, username)
		unpackDst = fmt.Sprintf("%s/src/github.com/%s/%s-%s", gopath, username, os.Getenv("COCOON_ID"), tagStr)
	}

	log.Info("Downloading cocoon repository with tag=%s, dst=%s", tagStr, downloadDst)
	filePath := fmt.Sprintf("%s/%s.tar.gz", downloadDst, os.Getenv("COCOON_ID"))
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

// Build go code and move binary to bin directory
func buildGoSource(sourceDir string) error {

	log.Infof("Building cocoon code source")

	output, err := exec.Command("/bin/sh", "-c", fmt.Sprintf("cd %s mkdir .build && go build -o .build/cc", sourceDir)).Output()
	if err != nil {
		return fmt.Errorf("Failed to install cocoon code. Output: %s. Error: %s", string(output), err)
	}

	log.Infof("Successfully built cocoon code")
	return nil
}

// Run an installed go coccon code
func startGo(ccRepoDir string) *exec.Cmd {
	log.Info("Building & Starting cocoon code")

	ccID := os.Getenv("COCOON_ID")
	build := fmt.Sprintf("cd %s && go build -o cc", ccRepoDir)
	run := fmt.Sprintf("./cc")

	return exec.Command("bash", "-c", strings.TrimSpace(fmt.Sprintf(`
		cd %s &&
		docker run --rm --name="%s" -v="$PWD:%s" ncodes/launch-go:latest bash -c "%s"
	`, ccRepoDir, ccID, ccRepoDir, strings.Join([]string{build, run}, " && "))))
}

// Start the recently installed cocoon code
func start(lang, ccRepoDir string) {
	switch lang {
	case "go":
		cmdCCode = startGo(ccRepoDir)
	}

	cmdCCReader, err := cmdCCode.StdoutPipe()
	if err != nil {
		log.Errorf("Failed to create StdoutPipe from cocoon code command. %s", err)
		return
	}

	err = cmdCCode.Start()
	if err != nil {
		log.Errorf("Failed to start cocoon code. %s", err)
		return
	}

	ccLog := logging.MustGetLogger("ccode")
	scanner := bufio.NewScanner(cmdCCReader)
	go func() {
		for scanner.Scan() {
			ccLog.Debug(scanner.Text())
		}
	}()

	done := make(chan error, 1)
	go func() {
		done <- cmdCCode.Wait()
	}()

	select {
	case err := <-done:
		if err != nil {
			log.Errorf("Cocoon code stopped with error = %s", err)
		} else {
			log.Info("Cocoon code has stopped gracefully")
		}
	}
}
