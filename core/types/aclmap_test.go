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

		Convey(".Eql", func() {

			Convey("Should return false since sizes are different", func() {
				x := NewACLMap(map[string]interface{}{
					"ledger1": map[string]string{
						"@identity_id": "allow",
						"cocoon_id":    "allow",
					},
				})
				y := NewACLMap(map[string]interface{}{
					"*": "allow",
					"ledger1": map[string]string{
						"@identity_id": "allow",
						"cocoon_id":    "allow",
					},
				})
				So(x.Eql(y), ShouldEqual, false)
			})

			Convey("Should return false since content is different", func() {
				x := NewACLMap(map[string]interface{}{
					"*": "deny",
					"ledger1": map[string]string{
						"@identity_id": "allow",
						"cocoon_id":    "allow",
					},
				})
				y := NewACLMap(map[string]interface{}{
					"*": "allow",
					"ledger1": map[string]string{
						"@identity_id": "allow",
						"cocoon_id":    "allow",
					},
				})
				So(x.Eql(y), ShouldEqual, false)
				z := NewACLMap(map[string]interface{}{
					"*": "deny",
					"ledger2": map[string]string{
						"@identity_id": "allow",
						"cocoon_id":    "allow",
					},
				})
				So(x.Eql(z), ShouldEqual, false)
			})

			Convey("Should return true since contents match", func() {
				x := NewACLMap(map[string]interface{}{
					"*": "deny",
					"ledger1": map[string]string{
						"@identity_id": "allow",
						"cocoon_id":    "allow",
					},
				})
				y := NewACLMap(map[string]interface{}{
					"*": "deny",
					"ledger1": map[string]string{
						"@identity_id": "allow",
						"cocoon_id":    "allow",
					},
				})
				So(x.Eql(y), ShouldEqual, true)
			})

			Convey("Should return true since contents match even in different order", func() {
				x := NewACLMap(map[string]interface{}{
					"ledger1": map[string]string{
						"cocoon_id":    "allow",
						"@identity_id": "allow",
					},
					"*": "deny",
				})
				y := NewACLMap(map[string]interface{}{
					"*": "deny",
					"ledger1": map[string]string{
						"@identity_id": "allow",
						"cocoon_id":    "allow",
					},
				})
				So(x.Eql(y), ShouldEqual, true)
			})
		})

		Convey(".Remove", func() {

			Convey("Should successfully remove an existing, non-actor specific rule", func() {
				aclMap := NewACLMap(map[string]interface{}{
					"*": "allow",
					"ledger1": map[string]string{
						"@identity_id": "allow",
						"cocoon_id":    "allow",
					},
				})
				aclMap.Remove("*")
				So(aclMap, ShouldResemble, ACLMap(map[string]interface{}{
					"ledger1": map[string]string{
						"@identity_id": "allow",
						"cocoon_id":    "allow",
					},
				}))
				So(aclMap, ShouldNotResemble, ACLMap(map[string]interface{}{
					"*": "allow",
					"ledger1": map[string]string{
						"@identity_id": "allow",
						"cocoon_id":    "allow",
					},
				}))
			})

			Convey("Should successfully remove an existing, actor specific rule and leaving actor rules intact", func() {
				aclMap := NewACLMap(map[string]interface{}{
					"*": "allow",
					"ledger1": map[string]string{
						"@identity_id": "allow",
						"cocoon_id":    "allow",
					},
				})
				aclMap.Remove("ledger1.cocoon_id")
				So(aclMap, ShouldResemble, ACLMap(map[string]interface{}{
					"*": "allow",
					"ledger1": map[string]string{
						"@identity_id": "allow",
					},
				}))
				So(aclMap, ShouldNotResemble, ACLMap(map[string]interface{}{
					"*": "allow",
					"ledger1": map[string]string{
						"@identity_id": "allow",
						"cocoon_id":    "allow",
					},
				}))
			})

			Convey("Should successfully remove all actor-specific rule if target is removed", func() {
				aclMap := NewACLMap(map[string]interface{}{
					"*": "allow",
					"ledger1": map[string]string{
						"@identity_id": "allow",
						"cocoon_id":    "allow",
					},
				})
				aclMap.Remove("ledger1")
				So(aclMap, ShouldResemble, ACLMap(map[string]interface{}{
					"*": "allow",
				}))
				aclMap = NewACLMap(map[string]interface{}{
					"*": "allow",
					"ledger1": map[string]string{
						"@identity_id": "allow",
						"cocoon_id":    "allow",
					},
				})
				aclMap.Remove("ledger1.@identity_id")
				aclMap.Remove("ledger1.cocoon_id")
				So(aclMap, ShouldResemble, ACLMap(map[string]interface{}{
					"*": "allow",
				}))
			})
		})
	})
}
