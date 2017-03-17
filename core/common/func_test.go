package common

import (
	"testing"
	"time"

	"fmt"

	"github.com/ncodes/cocoon/core/stubs/golang/proto"
	"github.com/ncodes/cocoon/core/types"
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

		Convey(".AwaitTxChanX", func() {
			Convey("Should return timeout error if channel did not return a value after 1 second", func() {
				ch := make(chan *proto.Tx)
				v, err := AwaitTxChanX(ch, 1*time.Second)
				So(v, ShouldBeNil)
				So(err, ShouldNotBeNil)
				So(err, ShouldResemble, types.ErrOperationTimeout)
			})
		})

		Convey(".ReRunOnError", func() {
			Convey("Should successfully rerun function", func() {
				runCount := 0
				err := ReRunOnError(func() error {
					runCount++
					if runCount != 3 {
						return fmt.Errorf("error")
					}
					return nil
				}, 3, nil)
				So(err, ShouldBeNil)
				So(runCount, ShouldEqual, 3)
			})

			Convey("Expects rerun to fail after max re-run limit is reached without success", func() {
				runCount := 0
				err := ReRunOnError(func() error {
					runCount++
					return fmt.Errorf("error")
				}, 3, nil)
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "error")
				So(runCount, ShouldEqual, 3)
			})

			Convey("Should successfully rerun function with delay", func() {
				runCount := 0
				delay := time.Millisecond * 100
				err := ReRunOnError(func() error {
					runCount++
					if runCount != 3 {
						return fmt.Errorf("error")
					}
					return nil
				}, 3, &delay)
				So(err, ShouldBeNil)
				So(runCount, ShouldEqual, 3)
			})
		})
	})
}
