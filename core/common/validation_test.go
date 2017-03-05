package common

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestValidation(t *testing.T) {
	Convey("TestValidation", t, func() {
		Convey(".ValidateDeployment", func() {
			Convey("should return error when url is not provided", func() {
				err := ValidateDeployment("", "", "")
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "url is required")
			})

			Convey("should return error when url is not a github repo url", func() {
				err := ValidateDeployment("http://google.com", "", "")
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "url is not a valid github repo url")
			})

			Convey("should return error when languages is not supported/unknown", func() {
				err := ValidateDeployment("https://github.com/user/repo", "xyz", "")
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "language is not supported. Expected one of these languages")
			})

			Convey("should return error when build params is not valid json", func() {
				err := ValidateDeployment("https://github.com/user/repo", "go", `{"key"}`)
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "build parameter is not valid json")
			})
		})
	})
}
