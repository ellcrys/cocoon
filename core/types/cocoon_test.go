package types

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestCocoon(t *testing.T) {
	Convey("Cocoon", t, func() {

		Convey(".Merge", func() {
			a := Cocoon{ID: "abc"}
			b := Cocoon{ID: "xyz"}
			err := a.Merge(b)
			So(err, ShouldBeNil)
			So(a.ID, ShouldEqual, "xyz")

			a = Cocoon{ID: "abc"}
			b = Cocoon{ID: ""}
			err = a.Merge(b)
			So(err, ShouldBeNil)
			So(a.ID, ShouldEqual, "abc")
		})

		Convey(".Difference", func() {
			a := Cocoon{ID: "abc", Memory: 100, NumSignatories: 2}
			b := Cocoon{ID: "abc"}
			diffs := a.Difference(b)
			So(len(diffs), ShouldEqual, 1)
			So(len(diffs[0]), ShouldEqual, 2)

			a = Cocoon{ID: "abc", Memory: 100, NumSignatories: 2}
			b = Cocoon{ID: "abc", Memory: 200}
			diffs = a.Difference(b)
			So(len(diffs), ShouldEqual, 1)
			So(len(diffs[0]), ShouldEqual, 2)

			a = Cocoon{ID: "abc", Memory: 100, NumSignatories: 2}
			b = Cocoon{}
			diffs = a.Difference(b)
			So(len(diffs), ShouldEqual, 1)
			So(len(diffs[0]), ShouldEqual, 3)

			a = Cocoon{ID: "abc", Memory: 100, NumSignatories: 2}
			b = Cocoon{ID: "abc", Memory: 100, NumSignatories: 2}
			diffs = a.Difference(b)
			So(len(diffs), ShouldEqual, 1)
			So(len(diffs), ShouldEqual, 1)
			So(diffs[0], ShouldBeNil)

			a = Cocoon{ID: "abc", Memory: 100}
			b = Cocoon{ID: "abc", Memory: 100, NumSignatories: 2}
			diffs = a.Difference(b)
			So(len(diffs), ShouldEqual, 1)
			So(len(diffs[0]), ShouldEqual, 1)
		})
	})
}
