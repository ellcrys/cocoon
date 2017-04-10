package acl

import (
	"testing"

	"github.com/ncodes/cocoon/core/types"
	. "github.com/smartystreets/goconvey/convey"
)

func TestInterpreter(t *testing.T) {
	Convey("Interpreter", t, func() {

		Convey(".Validate", func() {

			Convey("Should return nil if rule is empty", func() {
				i := NewInterpreter(map[string]interface{}{}, false)
				So(i.Validate(), ShouldBeNil)
			})

			Convey("Should return err if ledger rule value has unexpected type", func() {
				i := NewInterpreter(map[string]interface{}{"ledger1": 1}, false)
				errs := i.Validate()
				So(errs, ShouldNotBeEmpty)
				So(errs[0].Error(), ShouldEqual, "ledger1: invalid ledger value type. Expects string or map of strings")
			})

			Convey("Should return err if wildcard ledger rule value type is not string", func() {
				i := NewInterpreter(map[string]interface{}{"*": 1}, false)
				errs := i.Validate()
				So(errs, ShouldNotBeEmpty)
				So(errs[0].Error(), ShouldEqual, "*: invalid wildcard ledger value type. Expects string value")
			})

			Convey("Should return err if ledger rule contains invalid privileges", func() {
				i := NewInterpreter(map[string]interface{}{"ledger1": "something unknown"}, false)
				errs := i.Validate()
				So(len(errs), ShouldEqual, 2)
				So(errs[0].Error(), ShouldEqual, "ledger1: ledger contains an invalid privilege (something)")
				So(errs[1].Error(), ShouldEqual, "ledger1: ledger contains an invalid privilege (unknown)")
			})

			Convey("Should return error if legder rule contains invalid actor id", func() {
				i := NewInterpreter(map[string]interface{}{
					"ledger1": map[string]string{
						"": "allow",
					},
				}, false)
				errs := i.Validate()
				So(len(errs), ShouldEqual, 1)
				So(errs[0].Error(), ShouldEqual, "ledger1: invalid actor id. cocoon or identity id cannot be an empty string")
			})

			Convey("Should return error if legder rule actor's privileges contains invalid privilege", func() {
				i := NewInterpreter(map[string]interface{}{
					"ledger1": map[string]string{
						"@identity": "invalid",
					},
				}, false)
				errs := i.Validate()
				So(len(errs), ShouldEqual, 1)
				So(errs[0].Error(), ShouldEqual, "ledger1: ledger actor contains an invalid privilege (invalid)")
			})

			Convey("Should validate successfully if ledger rule type is valid and value is valid privilege", func() {
				i := NewInterpreter(map[string]interface{}{"ledger1": "allow deny"}, false)
				errs := i.Validate()
				So(len(errs), ShouldEqual, 0)
			})

		})

		Convey(".IsAllowed. Should successfully return expected results", func() {

			// rules, ledger name, actor id, operation, default policy, expected result
			var cases = [][]interface{}{
				[]interface{}{
					map[string]interface{}{}, "ledger1", "actor_id", types.TxPut, true, true,
				},
				[]interface{}{
					map[string]interface{}{}, "ledger1", "actor_id", types.TxPut, false, false,
				},
				[]interface{}{
					map[string]interface{}{
						"*": "allow deny",
					}, "ledger1", "actor_id", types.TxPut, false, false,
				},
				[]interface{}{
					map[string]interface{}{
						"*": "allow deny",
					}, "ledger1", "actor_id", types.TxPut, true, false,
				},
				[]interface{}{
					map[string]interface{}{
						"*": "deny",
					}, "ledger1", "actor_id", types.TxPut, true, false,
				},
				[]interface{}{
					map[string]interface{}{
						"*": "deny",
					}, "ledger1", "actor_id", types.TxGet, true, false,
				},
				[]interface{}{
					map[string]interface{}{
						"*": "allow",
					}, "ledger1", "actor_id", types.TxGet, true, true,
				},
				[]interface{}{
					map[string]interface{}{
						"*": "allow-put",
					}, "ledger1", "actor_id", types.TxGet, false, false,
				},
				[]interface{}{
					map[string]interface{}{
						"*": "allow-put",
					}, "ledger1", "actor_id", types.TxPut, false, true,
				},
				[]interface{}{
					map[string]interface{}{
						"*": "allow-put",
					}, "ledger1", "actor_id", types.TxGet, true, true,
				},
				[]interface{}{
					map[string]interface{}{
						"*": "allow-put",
					}, "ledger1", "actor_id", types.TxGet, false, false,
				},
				[]interface{}{
					map[string]interface{}{
						"*": "allow-put",
					}, "ledger1", "actor_id", types.TxPut, false, true,
				},
				[]interface{}{
					map[string]interface{}{
						"*": "allow-put deny-put",
					}, "ledger1", "actor_id", types.TxPut, false, false,
				},
				[]interface{}{
					map[string]interface{}{
						"*": "allow-put deny-put allow",
					}, "ledger1", "actor_id", types.TxPut, false, true,
				},
				[]interface{}{
					map[string]interface{}{
						"*":       "allow",
						"ledger1": "deny",
					}, "ledger1", "actor_id", types.TxPut, true, false,
				},
				[]interface{}{
					map[string]interface{}{
						"*":       "deny",
						"ledger1": "allow",
					}, "ledger1", "actor_id", types.TxPut, false, true,
				},
				[]interface{}{
					map[string]interface{}{
						"*":       "deny",
						"ledger2": "allow",
					}, "ledger1", "actor_id", types.TxPut, true, false,
				},
				[]interface{}{
					map[string]interface{}{
						"*": "allow",
						"ledger1": map[string]string{
							"actor1": "deny",
						},
					}, "ledger1", "actor1", types.TxPut, true, false,
				},
				[]interface{}{
					map[string]interface{}{
						"ledger1": map[string]string{
							"actor1": "deny",
						},
					}, "ledger1", "actor1", types.TxPut, true, false,
				},
				[]interface{}{
					map[string]interface{}{
						"ledger1": map[string]string{
							"actor1": "allow",
						},
					}, "ledger1", "actor1", types.TxPut, true, true,
				},
				[]interface{}{
					map[string]interface{}{
						"*": "deny",
						"ledger1": map[string]interface{}{
							"actor1":     "allow",
							"some_actor": "allow",
						},
					}, "ledger1", "actor1", types.TxGet, false, true,
				},
			}

			for _, c := range cases {
				i := NewInterpreter(c[0].(map[string]interface{}), c[4].(bool))
				expected := i.IsAllowed(c[1].(string), c[2].(string), c[3].(string))
				So(expected, ShouldEqual, c[5].(bool))
			}
		})
	})
}
