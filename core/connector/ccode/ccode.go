package ccode

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

var log = logging.MustGetLogger("connector")
var lastBuiltBinary = ""
var ccodeLang = ""
var cmdCCode *exec.Cmd

// GetJob returns the job details if
// it was previously deployed
func GetJob(jobName string) string {
	// not implemented
	return ""
}

// Install takes the cocoon code specified in
// the COCOON_CODE_URL environment variable,
// and installs it based on the language specified
// in COCOON_CODE_LANG
func Install(failed chan bool) {

	var ccID, ccURL, ccTag, ccLang string
	var err error

	log.Info("Ready to install cocoon code")

	// get cocoon code github link and language
	ccID = os.Getenv("COCOON_ID")
	ccURL = os.Getenv("COCOON_CODE_URL")
	ccTag = os.Getenv("COCOON_CODE_TAG")
	ccLang = os.Getenv("COCOON_CODE_LANG")

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
		if err = installFromGit(ccURL, ccTag, ccLang); err != nil {
			log.Error(err.Error())
			failed <- true
		}

		failed <- false
		return
	}

	log.Errorf("Cocoon url [%s] is not a github repo link", ccURL)
	failed <- true
}

// installFromGit fetchs cocoon code from git repo.
func installFromGit(url, tag, lang string) error {

	var prevJob, repoTarURL, downloadDst, unpackDst string
	var err error
	ccodeLang = lang

	// checks if job was previously deployed. find a job by the job name.
	prevJob = GetJob(os.Getenv("COCOON_ID"))
	if prevJob == "" {
		repoTarURL, err = others.GetGithubRepoRelease(url, tag)
		if err != nil {
			return fmt.Errorf("Failed to fetch release from github repo. %s", err)
		}
	} else {
		// repoID = prevRepoID
		// log.Debugf("cocoon was previously deployed. Using previous repo commit: %", prevRepoID)
		return fmt.Errorf("Job previously deployed. Redeployment not implemented yet")
	}

	tagStr := tag
	if tag == "" {
		tagStr = "latest"
	}

	// construct download directory
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
		return err
	}

	log.Info("Successfully downloaded cocoon code")
	log.Debugf("Unpacking cocoon code to %s", filePath)

	if err = os.MkdirAll(unpackDst, os.ModePerm); err != nil {
		return fmt.Errorf("Failed to create unpack directory. %s", err)
	}

	// unpack tarball
	cmd := "tar"
	args := []string{"-xf", filePath, "-C", unpackDst, "--strip-components", "1"}
	if err = exec.Command(cmd, args...).Run(); err != nil {
		return fmt.Errorf("Failed to unpack cocoon code repo tarball. %s", err)
	}

	log.Infof("Successfully unpacked cocoon code to %s", unpackDst)
	os.Remove(filePath)

	// build binary
	switch lang {
	case "go":
		log.Infof("Installing cocoon code")
		cmd := "go-echo-server"
		// output, err := exec.Command("/bin/sh", "-c", fmt.Sprintf("cd %s && go install", unpackDst)).Output()
		log.Debugf("Running command -> %s", cmd)
		output, err := exec.Command(cmd, "").Output()
		if err != nil {
			return fmt.Errorf("Failed to install cocoon code. Output: %s. Error: %s", string(output), err)
		}

		log.Infof("Successfully installed cocoon code")

		lastBuiltBinary = fmt.Sprintf("%s-%s", os.Getenv("COCOON_ID"), tagStr)

	default:
		return fmt.Errorf("Unsupported cocoon code language: %s", lang)
	}

	return nil
}

// Start the recently installed cocoon code
func Start() {

	switch ccodeLang {
	case "go":

		if lastBuiltBinary == "" {
			log.Error("Attempting to start cocoon code when binary has not been built. This is odd! Did you call Install() before calling this?")
		}

		cmdCCode := exec.Command(lastBuiltBinary)
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

		ccLog := logging.MustGetLogger(lastBuiltBinary)
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
				log.Errorf("Cocoon code stopped with error = %v", err)
			} else {
				log.Info("Cocoon code has stopped gracefully")
			}
		}
	}
}
