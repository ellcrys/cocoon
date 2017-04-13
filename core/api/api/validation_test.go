package api

import (
	"testing"

	"github.com/ellcrys/util"
	"github.com/ncodes/cocoon/core/types"
	. "github.com/smartystreets/goconvey/convey"
)

func TestValidation(t *testing.T) {
	Convey("Validation", t, func() {
		Convey(".ValidateCocoon", func() {
			Convey("should return expected errors", func() {

				err := ValidateCocoon(&types.Cocoon{})
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "id is required")

				err = ValidateCocoon(&types.Cocoon{
					ID: "some id",
				})
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "id is not a valid uuid")

				err = ValidateCocoon(&types.Cocoon{
					ID: util.UUID4(),
				})
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "url is required")

				err = ValidateCocoon(&types.Cocoon{
					ID:  util.UUID4(),
					URL: "http://google.com",
				})
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "url is not a valid github repo url")

				err = ValidateCocoon(&types.Cocoon{
					ID:  util.UUID4(),
					URL: "https://github.com/ncodes/cocoon-example-01",
				})
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "language is required")

				err = ValidateCocoon(&types.Cocoon{
					ID:       util.UUID4(),
					URL:      "https://github.com/ncodes/cocoon-example-01",
					Language: "c#",
				})
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "language is not supported. Expects one of these values [go]")

				err = ValidateCocoon(&types.Cocoon{
					ID:       util.UUID4(),
					URL:      "https://github.com/ncodes/cocoon-example-01",
					Language: "go",
				})
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "memory is required")

				err = ValidateCocoon(&types.Cocoon{
					ID:       util.UUID4(),
					URL:      "https://github.com/ncodes/cocoon-example-01",
					Language: "go",
					Memory:   "-1x",
				})
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "Memory value is not supported. Expects one of these values [512m 1g 2g]")

				err = ValidateCocoon(&types.Cocoon{
					ID:       util.UUID4(),
					URL:      "https://github.com/ncodes/cocoon-example-01",
					Language: "go",
					Memory:   "512m",
				})
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "CPU share is required")

				err = ValidateCocoon(&types.Cocoon{
					ID:        util.UUID4(),
					URL:       "https://github.com/ncodes/cocoon-example-01",
					Language:  "go",
					Memory:    "512m",
					CPUShares: "abc",
				})
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "CPU share value is not supported. Expects one of these values [1x 2x]")

				err = ValidateCocoon(&types.Cocoon{
					ID:             util.UUID4(),
					URL:            "https://github.com/ncodes/cocoon-example-01",
					Language:       "go",
					Memory:         "512m",
					CPUShares:      "1x",
					NumSignatories: 1,
					SigThreshold:   1,
					Signatories:    []string{"id1", "id2"},
				})
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "max signatories already added. You can't add more")
			})
		})

		Convey(".ValidateRelease", func() {

			err := ValidateRelease(&types.Release{})
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "id is required")

			err = ValidateRelease(&types.Release{
				ID: "some id",
			})
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "id is not a valid uuid")

			err = ValidateRelease(&types.Release{
				ID: util.UUID4(),
			})
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "cocoon id is required")

			err = ValidateRelease(&types.Release{
				ID:       util.UUID4(),
				CocoonID: "cocoon-123",
			})
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "url is required")

			err = ValidateRelease(&types.Release{
				ID:       util.UUID4(),
				CocoonID: "cocoon-123",
				URL:      "http://google.com",
			})
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "url is not a valid github repo url")

			err = ValidateRelease(&types.Release{
				ID:       util.UUID4(),
				CocoonID: "cocoon-123",
				URL:      "https://github.com/ncodes/cocoon-example-01",
			})
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "language is required")

			err = ValidateRelease(&types.Release{
				ID:       util.UUID4(),
				CocoonID: "cocoon-123",
				URL:      "https://github.com/ncodes/cocoon-example-01",
				Language: "abc",
			})
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "language is not supported")

			err = ValidateRelease(&types.Release{
				ID:         util.UUID4(),
				CocoonID:   "cocoon-123",
				URL:        "https://github.com/ncodes/cocoon-example-01",
				Language:   supportedLanguages[0],
				BuildParam: "non json",
			})
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "build parameter is not valid json")
		})

		Convey(".ValidateFirewall", func() {
			_, errs := ValidateFirewall("")
			So(len(errs), ShouldEqual, 1)
			So(errs[0].Error(), ShouldEqual, "empty string passed")

			_, errs = ValidateFirewall(123)
			So(len(errs), ShouldEqual, 1)
			So(errs[0].Error(), ShouldEqual, "invalid type. expects a json string or a slice of map")

			_, errs = ValidateFirewall(`abc`)
			So(len(errs), ShouldEqual, 1)
			So(errs[0].Error(), ShouldEqual, "malformed json")

			_, errs = ValidateFirewall(`{}`)
			So(len(errs), ShouldEqual, 1)
			So(errs[0].Error(), ShouldEqual, "malformed json")

			_, errs = ValidateFirewall(`[{}]`)
			So(len(errs), ShouldEqual, 2)
			So(errs[0].Error(), ShouldEqual, "rule 0: destination is required")
			So(errs[1].Error(), ShouldEqual, "rule 0: port is required")

			_, errs = ValidateFirewall(`[{ "destination": "0.0.0.0.0" }]`)
			So(len(errs), ShouldEqual, 2)
			So(errs[0].Error(), ShouldEqual, "rule 0: destination is not a valid IP or host")
			So(errs[1].Error(), ShouldEqual, "rule 0: port is required")

			_, errs = ValidateFirewall(`[{ "destination": "http://google.com" }]`)
			So(len(errs), ShouldEqual, 2)
			So(errs[0].Error(), ShouldEqual, "rule 0: destination is not a valid IP or host")
			So(errs[1].Error(), ShouldEqual, "rule 0: port is required")

			_, errs = ValidateFirewall(`[{ "destination": "google.com", "protocol": "icmp" }]`)
			So(len(errs), ShouldEqual, 2)
			So(errs[0].Error(), ShouldEqual, "rule 0: port is required")
			So(errs[1].Error(), ShouldEqual, "rule 0: invalid protocol")

			_, errs = ValidateFirewall(`[{ "destination": "0.0.0.0", "port": "3000" }]`)
			So(len(errs), ShouldEqual, 0)
		})
	})
}
