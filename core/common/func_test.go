package common

import (
	"errors"
	"testing"
	"time"

	"fmt"

	"github.com/ncodes/cocoon/core/runtime/golang/proto"
	"github.com/ncodes/cocoon/core/types"
	. "github.com/smartystreets/goconvey/convey"
)

func TestFunc(t *testing.T) {
	Convey("Func", t, func() {
		Convey(".GetRPCErrDesc", func() {
			Convey("Should remove rpc error = 2 from error", func() {
				bs := GetRPCErrDesc(errors.New("rpc error: code = 2 desc = something bad happened"))
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

		Convey(".CompareErr", func() {
			Convey("Should successfully match both errors", func() {
				So(CompareErr(errors.New("a"), errors.New("a")), ShouldEqual, 0)
			})
			Convey("Should return non zero (-1 or 1) if errors don't match", func() {
				So(CompareErr(errors.New("a"), errors.New("b")), ShouldEqual, -1)
				So(CompareErr(errors.New("b"), errors.New("a")), ShouldEqual, 1)
			})
		})

		Convey(".CapitalizeString", func() {
			Convey("Should successfully capitalize strings", func() {
				cases := [][]string{
					[]string{"the people are smiling", "The people are smiling"},
					[]string{"the people are smiling. they love the President", "The people are smiling. They love the President"},
				}
				for _, c := range cases {
					So(CapitalizeString(c[0]), ShouldEqual, c[1])
				}
			})
		})
	})
}
