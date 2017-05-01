package types

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestCocoon(t *testing.T) {
	Convey("Cocoon", t, func() {
		Convey(".Difference", func() {
			// a := Cocoon{ID: "abc", Memory: 100, NumSignatories: 2}
			// b := Cocoon{ID: "abc"}
			// isDiff, diffs := a.Difference(b)
			// So(isDiff, ShouldEqual, true)
			// pretty.Println(diffs)

			// a = Cocoon{ID: "abc", Memory: 100, NumSignatories: 2}
			// b = Cocoon{ID: "abc", Memory: 200}
			// So(a.Difference(b), ShouldEqual, true)

			// a = Cocoon{ID: "abc", Memory: 100, NumSignatories: 2}
			// b = Cocoon{}
			// So(a.Difference(b), ShouldEqual, true)

			// a = Cocoon{ID: "abc", Memory: 100, NumSignatories: 2}
			// b = Cocoon{ID: "abc", Memory: 100, NumSignatories: 2}
			// So(a.Difference(b), ShouldEqual, false)

			// a = Cocoon{ID: "abc", Memory: 100}
			// b = Cocoon{ID: "abc", Memory: 100, NumSignatories: 2}
			// So(a.Difference(b), ShouldEqual, true)
		})
	})
}
