package archiver

import (
	"testing"

	"github.com/ellcrys/util"
	. "github.com/smartystreets/goconvey/convey"
)

func TestGitObject(t *testing.T) {
	Convey("GitObject", t, func() {
		Convey(".Read", func() {
			Convey("Should return error if repository is not a valid github repository", func() {
				o := NewGitObject("https://gitbla.com/ncodes/safebuffer", "")
				err := o.Read(nil)
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "repo url is invalid")
			})

			Convey("Should return error if repo does not exists", func() {
				o := NewGitObject("https://github.com/ncodes/safeblabla", "")
				err := o.Read(nil)
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "failed to get release tarball")
				So(err.Error(), ShouldContainSubstring, "Not Found")
			})

			Convey("Should return error if repo release does not exists", func() {
				o := NewGitObject("https://github.com/ncodes/safebuffer", "unknown_bla_bla")
				err := o.Read(nil)
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "failed to get release tarball")
				So(err.Error(), ShouldContainSubstring, "Not Found")
			})

			Convey("Should return error if repo release tag does not exists", func() {
				o := NewGitObject("https://github.com/ncodes/safebuffer", util.RandString(40))
				err := o.Read(func(b []byte) error { return nil })
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "failed to fetch tarball")
			})

			Convey("Should successfully read repo", func() {
				copiedCount := 0
				o := NewGitObject("https://github.com/ncodes/safebuffer", "bf465d911d23075ab653f42a3a0d315584383612")
				err := o.Read(func(b []byte) error {
					copiedCount = copiedCount + len(b)
					return nil
				})
				So(err, ShouldBeNil)
				So(copiedCount, ShouldBeGreaterThan, 0)
			})
		})
	})
}
