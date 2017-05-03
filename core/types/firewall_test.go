package types

import (
	"testing"

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

	})
}
