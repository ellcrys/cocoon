package common

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestFunc(t *testing.T) {
	Convey("Func", t, func() {
		Convey("StripRPCErrorPrefix", func() {
			Convey("Should remove rpc error = 2 from error", func() {
				bs := StripRPCErrorPrefix([]byte("rpc error: code = 2 desc = something bad happened"))
				So(string(bs), ShouldEqual, "something bad happened")
			})
		})
	})
}
