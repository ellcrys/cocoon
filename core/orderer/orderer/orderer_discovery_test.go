package orderer

import (
	"testing"
	"time"

	"os"

	"github.com/ellcrys/util"
	. "github.com/smartystreets/goconvey/convey"
)

func TestOrdererDiscovery(t *testing.T) {
	Convey("OrdererDiscovery", t, func() {
		Convey(".NewDiscovery", func() {
			Convey("Should return error if unable to connector to consul", func() {
				os.Setenv("CONSUL_ADDR", "localhost:4444")
				_, err := NewDiscovery()
				So(err, ShouldNotBeNil)
				os.Setenv("CONSUL_ADDR", "")
			})

			Convey("Should successfully ping consul and return no error", func() {
				_, err := NewDiscovery()
				So(err, ShouldBeNil)
			})
		})

		Convey(".Discover", func() {
			DiscoveryInterval = 2 * time.Second
			discovery, err := NewDiscovery()
			So(err, ShouldBeNil)

			Convey("Should return no error", func() {
				err := discovery.addOrdererService("1.1.1.1", 7000)
				So(err, ShouldBeNil)
				go discovery.Discover()
				time.Sleep(1 * time.Second)

				Convey(".GetAddrs", func() {
					Convey("Should return a single addr", func() {
						addr := discovery.GetAddrs()
						util.Printify(addr)
						So(len(addr), ShouldEqual, 1)
						So(addr[0], ShouldEqual, "1.1.1.1:7000")
					})
				})
			})
		})
	})
}
