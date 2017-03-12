package cocoon

import (
	"testing"

	"github.com/ncodes/cocoon/core/types/client"
	. "github.com/smartystreets/goconvey/convey"
)

func TestCocoon(t *testing.T) {
	Convey("Cocoon", t, func() {
		Convey(".validateCreateCocoon", func() {
			Convey("should return expected errors", func() {

				err := validateCreateCocoon(&client.Cocoon{})
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "url is required")

				err = validateCreateCocoon(&client.Cocoon{
					URL: "http://google.com",
				})
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "url is not a valid github repo url")

				err = validateCreateCocoon(&client.Cocoon{
					URL: "https://github.com/ncodes/cocoon-example-01",
				})
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "language is required")

				err = validateCreateCocoon(&client.Cocoon{
					URL:      "https://github.com/ncodes/cocoon-example-01",
					Language: "c#",
				})
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "language is not supported. Expects one of these values [go]")

				err = validateCreateCocoon(&client.Cocoon{
					URL:      "https://github.com/ncodes/cocoon-example-01",
					Language: "go",
				})
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "memory is required")

				err = validateCreateCocoon(&client.Cocoon{
					URL:      "https://github.com/ncodes/cocoon-example-01",
					Language: "go",
					Memory:   "-1x",
				})
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "Memory value is not supported. Expects one of these values [512m 1g 2g]")

				err = validateCreateCocoon(&client.Cocoon{
					URL:      "https://github.com/ncodes/cocoon-example-01",
					Language: "go",
					Memory:   "512m",
				})
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "CPU share is required")

				err = validateCreateCocoon(&client.Cocoon{
					URL:      "https://github.com/ncodes/cocoon-example-01",
					Language: "go",
					Memory:   "512m",
					CPUShare: "abc",
				})
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "CPU share value is not supported. Expects one of these values [1x 2x]")
			})
		})
	})
}
