package types

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestRelease(t *testing.T) {
	Convey("Release", t, func() {
		Convey(".Difference", func() {

			a := Release{
				ACL: NewACLMap(map[string]interface{}{
					"*": "deny",
				}),
				Firewall: Firewall([]FirewallRule{
					FirewallRule{
						Destination:     "127.0.0.1",
						DestinationPort: "8080",
						Protocol:        "tcp",
					},
				}),
				Env: Env(map[string]string{
					"ABC": "xyz",
				}),
			}
			diff := a.Difference(a)
			So(len(diff), ShouldEqual, 1)
			So(diff[0], ShouldBeNil)

			a = Release{
				ACL: NewACLMap(map[string]interface{}{
					"*": "deny",
				}),
				Firewall: Firewall([]FirewallRule{
					FirewallRule{
						Destination:     "127.0.0.1",
						DestinationPort: "8080",
						Protocol:        "tcp",
					},
				}),
				Env: Env(map[string]string{
					"ABC": "www",
				}),
			}
			b := Release{
				ACL: NewACLMap(map[string]interface{}{
					"*": "deny",
				}),
				Firewall: Firewall([]FirewallRule{
					FirewallRule{
						Destination:     "127.0.0.1",
						DestinationPort: "8080",
						Protocol:        "tcp",
					},
				}),
				Env: Env(map[string]string{
					"ABC": "xyz",
				}),
			}
			diff2 := a.Difference(b)
			So(len(diff2), ShouldEqual, 1)
			So(diff2[0], ShouldNotBeNil)

			a = Release{
				ACL: NewACLMap(map[string]interface{}{
					"*": "deny",
				}),
				Firewall: Firewall([]FirewallRule{
					FirewallRule{
						Destination:     "127.0.0.1",
						DestinationPort: "8080",
						Protocol:        "tcp",
					},
				}),
				Env: Env(map[string]string{
					"ABC": "xyz",
				}),
			}
			b = Release{
				ACL: NewACLMap(map[string]interface{}{
					"*": "deny",
				}),
				Firewall: Firewall([]FirewallRule{
					FirewallRule{
						Destination:     "127.0.0.1",
						DestinationPort: "8080",
						Protocol:        "tcp",
					},
					FirewallRule{
						Destination:     "127.0.0.1",
						DestinationPort: "5000",
						Protocol:        "tcp",
					},
				}),
				Env: Env(map[string]string{
					"ABC": "xy",
				}),
			}

			diff3 := a.Difference(b)
			So(len(diff3), ShouldEqual, 1)
			So(len(diff3[0]), ShouldEqual, 2)
		})

		Convey(".Merge", func() {
			a := Release{CocoonID: "abc"}
			b := Release{CocoonID: "xyz"}
			err := a.Merge(b)
			So(err, ShouldBeNil)
			So(a.CocoonID, ShouldEqual, "xyz")

			a = Release{CocoonID: "abc"}
			b = Release{CocoonID: ""}
			err = a.Merge(b)
			So(err, ShouldBeNil)
			So(a.CocoonID, ShouldEqual, "abc")
		})
	})
}
