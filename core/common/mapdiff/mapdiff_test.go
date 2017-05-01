package mapdiff

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestMapDiff(t *testing.T) {
	Convey("MapDiff", t, func() {

		Convey(".Difference", func() {
			mapA := Map{Value: map[string]interface{}{"age": 30}, Name: "map_A"}
			mapB := Map{Value: map[string]interface{}{"age": 31}, Name: "map_B"}
			s := NewMapDiff(mapA, mapB)
			So(s.Difference(), ShouldResemble, []bool{true})

			mapA = Map{Value: map[string]interface{}{"age": 30}, Name: "map_A"}
			mapB = Map{Value: map[string]interface{}{"age": 30}, Name: "map_B"}
			s = NewMapDiff(mapA, mapB)
			So(s.Difference(), ShouldResemble, []bool{false})

			mapA = Map{Value: map[string]interface{}{"age": 30}, Name: "map_A"}
			mapB = Map{Value: map[string]interface{}{"age": 30}, Name: "map_B"}
			mapC := Map{Value: map[string]interface{}{"age": 31}, Name: "map_B"}
			s = NewMapDiff(mapA, mapB, mapC)
			So(s.Difference(), ShouldResemble, []bool{false, true})
		})

		Convey(".Diff", func() {

			Convey("Should return DiffValue when the shortest map is missing a field on the longest map", func() {
				mapA := Map{Value: map[string]interface{}{"age": 30}, Name: "map_A"}
				mapB := Map{Value: map[string]interface{}{"age": 30, "sex": "m"}, Name: "map_B"}
				s := NewMapDiff(mapA, mapB)
				So(s.Diff(), ShouldResemble, [][]DiffValue{
					[]DiffValue{{
						LongestMap:    mapB,
						AffectedField: []interface{}{map[string]interface{}{"sex": "m"}},
						Type:          DiffTypeMissing,
						AffectedMap:   mapA,
					}},
				})
			})

			Convey("Should return a nil value if a map is not different from the main map", func() {
				mapA := Map{Value: map[string]interface{}{"age": 30, "sex": "m"}, Name: "map_A"}
				mapB := Map{Value: map[string]interface{}{"age": 30, "sex": "m"}, Name: "map_B"}
				s := NewMapDiff(mapA, mapB)
				actual := s.Diff()
				So(len(actual), ShouldEqual, 1)
				So(actual[0], ShouldBeNil)
			})

			Convey("Should return expected result length for errors relating to missing or different field values", func() {
				mapA := Map{Value: map[string]interface{}{"age": 30, "sex": "f"}, Name: "map_A"}
				mapB := Map{Value: map[string]interface{}{"age": 30, "sex": "m"}, Name: "map_B"}
				s := NewMapDiff(mapA, mapB)
				expected := [][]DiffValue{
					[]DiffValue{{
						LongestMap:    mapA,
						Type:          DiffTypeDifferent,
						AffectedField: []interface{}{map[string]interface{}{"sex": "f"}},
						AffectedMap:   mapB,
					}},
				}
				actual := s.Diff()
				So(actual, ShouldResemble, expected)

				mapA = Map{Value: map[string]interface{}{"age": 30, "sex": "f", "mode": "warrior"}, Name: "map_A"}
				mapB = Map{Value: map[string]interface{}{"age": 30, "sex": "m"}, Name: "map_B"}
				mapC := Map{Value: map[string]interface{}{"age": 40, "sex": "m"}, Name: "map_C"}
				s = NewMapDiff(mapA, mapB, mapC)
				actual = s.Diff()
				So(len(actual), ShouldEqual, 2)
				So(len(actual[0]), ShouldEqual, 2)
				So(len(actual[1]), ShouldEqual, 3)
			})
		})
	})
}
