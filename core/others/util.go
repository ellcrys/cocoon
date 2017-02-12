package others

import (
	"regexp"
)

// IsGithubRepoURL checks whether a url is a valid github repo url.
func IsGithubRepoURL(link string) bool {
	return regexp.MustCompile("https?://github.com/.*/.*").Match([]byte(link))
}
