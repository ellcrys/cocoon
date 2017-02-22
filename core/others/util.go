package others

import (
	"fmt"
	"os"
	"regexp"
	"time"

	"net/http"

	"github.com/ellcrys/util"
	"github.com/franela/goreq"
	"github.com/goware/urlx"
)

func init() {
	goreq.SetConnectTimeout(10 * time.Second)
}

// IsGithubRepoURL checks whether a url is a valid github repo url.
func IsGithubRepoURL(url string) bool {
	return regexp.MustCompile("^https?://github.com/.*/.*$").Match([]byte(url))
}

// GithubGetLatestCommitID fetches the most recent commit
// of a github repo and returns the sha value
func GithubGetLatestCommitID(url string) (string, error) {
	if !IsGithubRepoURL(url) {
		return "", fmt.Errorf("invalid github repo url")
	}

	u, _ := urlx.Parse(url)
	res, err := goreq.Request{
		Uri: fmt.Sprintf("https://api.github.com/repos%s/commits", u.Path),
	}.Do()

	if err != nil {
		return "", fmt.Errorf("failed to fetch repo commits. %s", err)
	}

	body, _ := res.Body.ToString()
	if res.StatusCode != 200 {
		return "", fmt.Errorf("failed to fetch repo. %s", body)
	}

	var commits []map[string]interface{}
	if err = util.FromJSON([]byte(body), &commits); err != nil {
		return "", fmt.Errorf("malformed result from github. %s", body)
	}

	return commits[0]["sha"].(string), nil
}

// GetGithubRepoRelease returns the tarball
// url of a github repository release by tag.
// Fetches the latest release if tag is not spcified
func GetGithubRepoRelease(repoURL, tag string) (string, error) {
	var endpoint string
	var err error

	if !IsGithubRepoURL(repoURL) {
		return "", fmt.Errorf("invalid github repo url")
	}

	u, _ := urlx.Parse(repoURL)
	endpoint = fmt.Sprintf("https://api.github.com/repos%s/releases/tags/%s", u.Path, tag)
	if tag == "" {
		endpoint = fmt.Sprintf("https://api.github.com/repos%s/releases/latest", u.Path)
	}

	res, err := goreq.Request{
		Uri: endpoint,
	}.Do()

	if err != nil {
		return "", fmt.Errorf("failed to get repo release. %s", err)
	}

	body, _ := res.Body.ToString()
	if res.StatusCode != 200 {
		return "", fmt.Errorf("failed to fetch repo release. %s", body)
	}

	var release map[string]interface{}
	if err = util.FromJSON([]byte(body), &release); err != nil {
		return "", fmt.Errorf("malformed result from github. %s", body)
	}

	return release["tarball_url"].(string), nil
}

// DownloadFile downloads a remote file to a destination
func DownloadFile(url, destFilename string, cb func([]byte)) error {

	var err, writeErr error

	f, err := os.Create(destFilename)
	if err != nil {
		return fmt.Errorf("%s", err)
	}
	defer f.Close()

	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to download remote resource. %s", err)
	}

	defer resp.Body.Close()

	util.ReadReader(resp.Body, 32000, func(err error, buf []byte, done bool) bool {
		if !done {
			if _, e := f.Write(buf); e != nil {
				writeErr = e
				return false
			}
		}
		cb(buf)
		return true
	})

	f.Sync()
	return nil
}
