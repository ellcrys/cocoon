package util

import (
	"os"
	"testing"

	"github.com/ellcrys/util"
	. "github.com/smartystreets/goconvey/convey"
)

func TestUtil(t *testing.T) {

	Convey("Util", t, func() {
		Convey(".IsGithubRepoURL", func() {
			Convey("should succeed because url is a github repo url", func() {
				So(IsGithubRepoURL("https://github.com/ncodes/cocoon"), ShouldEqual, true)
			})

			Convey("should fail because url is a github repo url", func() {
				So(IsGithubRepoURL("https://gitlab.com/ncodes/cocoon"), ShouldEqual, false)
			})
		})

		Convey(".GithubGetLatestCommitID", func() {
			Convey("should return error if github repo url is invalid", func() {
				_, err := GithubGetLatestCommitID("http://githubs.com/user/repo")
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "invalid github repo url")
			})

			Convey("should successfully return repo sha hash", func() {
				sha, err := GithubGetLatestCommitID("http://github.com/ncodes/cocoon")
				So(err, ShouldBeNil)
				So(len(sha), ShouldEqual, 40)
			})
		})

		Convey(".GetGithubRepoRelease", func() {
			Convey("should return error if github repo url is invalid", func() {
				_, err := GetGithubRepoRelease("http://githubs.com/user/repo", "tag")
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "invalid github repo url")
			})

			Convey("should return error release is not found", func() {
				_, err := GetGithubRepoRelease("http://github.com/ncodes/cocoon-example-01", "unknown-tag")
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "Not Found")
			})

			Convey("should successfully return latest release tarball", func() {
				tarURL, err := GetGithubRepoRelease("http://github.com/ncodes/cocoon-example-01", "")
				So(err, ShouldBeNil)
				So(tarURL, ShouldContainSubstring, "tarball")
			})
		})

		Convey(".DownloadFile", func() {
			Convey("should successfully download a file", func() {
				remoteURL := "https://api.github.com/repos/ncodes/cocoon-example-01/tarball/v0.0.2"
				destFile := "/tmp/" + util.RandString(5) + ".tar.gz"
				defer os.Remove(destFile)
				DownloadFile(remoteURL, destFile, func(b []byte) {})
				_, err := os.Stat(destFile)
				So(err, ShouldNotEqual, os.ErrNotExist)
			})
		})

		Convey(".IsGithubCommitID", func() {
			Convey("should return true", func() {
				So(IsGithubCommitID("351e11dac558a764ba83f89c6598151d2dbaf904"), ShouldEqual, true)
			})
			Convey("should return false", func() {
				So(IsGithubCommitID("351e11dac558a764ba83"), ShouldEqual, false)
			})
		})

		Convey(".IsExistingGithubRepo", func() {
			Convey("should return true", func() {
				So(IsExistingGithubRepo("https://github.com/ncodes/cocoon-example-01"), ShouldEqual, true)
			})
			Convey("should return false", func() {
				So(IsExistingGithubRepo("https://github.com/ncodes/cocoon-example-10"), ShouldEqual, false)
			})
		})

		Convey(".IsValidGithubRelease", func() {
			Convey("should return true", func() {
				So(IsValidGithubRelease("https://github.com/ncodes/cocoon-example-01", "v0.0.2"), ShouldEqual, true)
			})
			Convey("should return false", func() {
				So(IsValidGithubRelease("https://github.com/ncodes/cocoon-example-01", "unknown"), ShouldEqual, false)
			})
		})

		Convey(".IsValidGithubCommitID", func() {
			Convey("should return true", func() {
				So(IsValidGithubCommitID("https://github.com/ncodes/cocoon-example-01", "cdac5dbe22fb5ddc21850f3fd91d75422f712a4a"), ShouldEqual, true)
			})
			Convey("should return false", func() {
				So(IsValidGithubCommitID("https://github.com/ncodes/cocoon-example-01", "unknown"), ShouldEqual, false)
			})
		})
	})
}
