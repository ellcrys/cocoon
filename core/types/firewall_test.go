package types

import (
	"testing"

	"github.com/ncodes/cocoon/core/api/api/proto_api"
	. "github.com/smartystreets/goconvey/convey"
)

func TestFirewall(t *testing.T) {
	Convey("Firewall", t, func() {
		Convey(".Eql", func() {
			Convey("Should return false if compared firewall does not match", func() {
				fw := Firewall([]FirewallRule{
					{Destination: "google.com", DestinationPort: "80", Protocol: "tcp"},
					{Destination: "facebool.com", DestinationPort: "80", Protocol: "tcp"},
				})
				fw2 := Firewall([]FirewallRule{
					{Destination: "google.com", DestinationPort: "80", Protocol: "tcp"},
					{Destination: "facebook.com", DestinationPort: "80", Protocol: "tcp"},
				})
				fw3 := Firewall([]FirewallRule{
					{Destination: "google.com", DestinationPort: "80", Protocol: "tcp"},
				})
				fw4 := Firewall([]FirewallRule{
					{Destination: "google.com", DestinationPort: "80", Protocol: "udp"},
				})
				So(fw.Eql(fw2), ShouldEqual, false)
				So(fw.Eql(fw3), ShouldEqual, false)
				So(fw3.Eql(fw4), ShouldEqual, false)
			})

			Convey("Should return true if compared firewall match", func() {
				fw := Firewall([]FirewallRule{
					{Destination: "google.com", DestinationPort: "80", Protocol: "tcp"},
					{Destination: "facebook.com", DestinationPort: "80", Protocol: "tcp"},
				})
				fw2 := Firewall([]FirewallRule{
					{Destination: "google.com", DestinationPort: "80", Protocol: "tcp"},
					{Destination: "facebook.com", DestinationPort: "80", Protocol: "tcp"},
				})
				So(fw.Eql(fw2), ShouldEqual, true)
			})
		})

		Convey(".DeDup", func() {
			Convey("Should return a de-duplicated firewall", func() {
				fw := Firewall([]FirewallRule{
					{Destination: "google.com", DestinationPort: "80", Protocol: "tcp"},
					{Destination: "google.com", DestinationPort: "80", Protocol: "tcp"},
				})
				expected := fw.DeDup()
				So(len(expected), ShouldEqual, 1)

				fw = Firewall([]FirewallRule{
					{Destination: "google.com", DestinationPort: "80", Protocol: "tcp"},
					{Destination: "nairaland.com", DestinationPort: "80", Protocol: "tcp"},
				})
				expected = fw.DeDup()
				So(len(expected), ShouldEqual, 2)
			})
		})

		Convey(".CopyFirewall", func() {

			Convey("Should successfully copy a Firewall object", func() {
				x := Firewall([]FirewallRule{
					{Destination: "127.0.0.1", DestinationPort: "3333", Protocol: "tcp"},
					{Destination: "127.0.0.1", DestinationPort: "3334", Protocol: "tcp"},
				})
				actual := CopyFirewall(x)
				So(len(actual), ShouldEqual, 2)
				So(actual, ShouldResemble, x)
			})

			Convey("Should successfully copy a slice of proto_api.FirewallRule object", func() {
				x := []proto_api.FirewallRule{
					{Destination: "127.0.0.1", DestinationPort: "3333", Protocol: "tcp"},
					{Destination: "127.0.0.1", DestinationPort: "3334", Protocol: "tcp"},
				}
				actual := CopyFirewall(x)
				So(len(actual), ShouldEqual, 2)
				So(actual[0].Destination, ShouldEqual, x[0].Destination)
				So(actual[0].DestinationPort, ShouldEqual, x[0].DestinationPort)
				So(actual[0].Protocol, ShouldEqual, x[0].Protocol)
				So(actual[1].Destination, ShouldEqual, x[1].Destination)
				So(actual[1].DestinationPort, ShouldEqual, x[1].DestinationPort)
				So(actual[1].Protocol, ShouldEqual, x[1].Protocol)
			})
		})
	})
}
