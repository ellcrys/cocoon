package common

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestFunc(t *testing.T) {
	Convey("Func", t, func() {
		Convey(".StripRPCErrorPrefix", func() {
			Convey("Should remove rpc error = 2 from error", func() {
				bs := StripRPCErrorPrefix([]byte("rpc error: code = 2 desc = something bad happened"))
				So(string(bs), ShouldEqual, "something bad happened")
			})
		})

		Convey(".IsValidResName", func() {

			cases := [][]interface{}{
				[]interface{}{"lord.luggard", false},
				[]interface{}{"lord_luggard", true},
				[]interface{}{"lord-luggard", false},
				[]interface{}{"lordluggard", true},
			}

			for _, c := range cases {
				So(IsValidResName(c[0].(string)), ShouldEqual, c[1].(bool))
			}
		})
	})
}
