package others

import (
	"testing"

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
	})
}
