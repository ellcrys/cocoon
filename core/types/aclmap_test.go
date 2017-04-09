package types

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestACLMap(t *testing.T) {
	Convey("ACLMap", t, func() {
		Convey(".NewACLMap", func() {
			Convey("Should successfully create a nil ACLMap object", func() {
				aclMap := NewACLMap(nil)
				So(aclMap, ShouldBeNil)
			})

			Convey("Should successfully create an ACLMap object with default values", func() {
				aclMap := NewACLMap(map[string]interface{}{
					"*": "allow",
				})
				So(aclMap, ShouldNotBeNil)
			})
		})
		Convey(".Add", func() {
			aclMap := NewACLMap(map[string]interface{}{
				"*": "allow",
			})

			Convey("Should return error if target format is invalid", func() {
				err := aclMap.Add("a.b.c", "allow")
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "target format is invalid")
			})

			Convey("Should successfully add a new target rule with privileges", func() {
				err := aclMap.Add("ledger1", "allow")
				So(err, ShouldBeNil)
				So(aclMap, ShouldResemble, ACLMap(map[string]interface{}{
					"*":       "allow",
					"ledger1": "allow",
				}))
			})

			Convey("Should successfully overwrite target rule with privileges", func() {
				err := aclMap.Add("*", "allow,deny")
				So(err, ShouldBeNil)
				So(aclMap, ShouldResemble, ACLMap(map[string]interface{}{
					"*": "allow,deny",
				}))
			})

			Convey("Should successfully add a new actor specific privilege", func() {
				err := aclMap.Add("ledger1.cocoon_id", "allow")
				So(err, ShouldBeNil)
				So(aclMap, ShouldResemble, ACLMap(map[string]interface{}{
					"*": "allow",
					"ledger1": map[string]string{
						"cocoon_id": "allow",
					},
				}))
				err = aclMap.Add("ledger1.@identity_id", "allow")
				So(err, ShouldBeNil)
				So(aclMap, ShouldResemble, ACLMap(map[string]interface{}{
					"*": "allow",
					"ledger1": map[string]string{
						"cocoon_id":    "allow",
						"@identity_id": "allow",
					},
				}))
			})
		})
	})
}
