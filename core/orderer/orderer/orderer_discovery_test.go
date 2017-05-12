package orderer

import (
	"testing"
	"time"

	"os"

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

			Convey("Tests", func() {
				err := discovery.addOrdererService("1.1.1.1", 7000)
				So(err, ShouldBeNil)
				go discovery.Discover()
				time.Sleep(1 * time.Second)

				Convey(".GetAddrs", func() {
					Convey("Should return a single addr", func() {
						addr := discovery.GetAddrs()
						So(len(addr), ShouldEqual, 1)
					})
				})

				Convey(".Add", func() {
					Convey("Should successfully add new address", func() {
						discovery.Add("1.1.1.2:1000")
						So(len(discovery.orderersAddr), ShouldEqual, 2)
					})
				})

				Convey(".Len", func() {
					Convey("Should successfully return expected length", func() {
						So(discovery.Len(), ShouldEqual, 1)
						discovery.Add("1.1.1.2:1000")
						So(discovery.Len(), ShouldEqual, 2)
					})
				})

				Convey(".GetGRPConn", func() {
					Convey("Should return error if no address is available", func() {
						discovery, err := NewDiscovery()
						_, err = discovery.GetGRPConn()
						So(err, ShouldNotBeNil)
						So(err.Error(), ShouldEqual, "no known orderer address")
					})

					Convey("Should successfully return a client", func() {
						c, err := discovery.GetGRPConn()
						So(err, ShouldBeNil)
						So(c, ShouldNotBeNil)
					})
				})
			})
		})
	})
}
