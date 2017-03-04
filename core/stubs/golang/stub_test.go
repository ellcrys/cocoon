package golang

import (
	"testing"

	"github.com/ncodes/cocoon/core/ledgerchain/types"
	. "github.com/smartystreets/goconvey/convey"
)

func TestStub(t *testing.T) {
	Convey("GoStub", t, func() {

		Convey("GetGlobalLedgerName", func() {
			Convey("Should return the expected value set in types.GetGlobalLedgerName()", func() {
				So(GetGlobalLedgerName(), ShouldEqual, types.GetGlobalLedgerName())
			})
		})

		Convey("stripRPCErrorPrefix", func() {
			Convey("Should remove rpc error = 2 from error", func() {
				bs := stripRPCErrorPrefix([]byte("rpc error: code = 2 desc = something bad happened"))
				So(string(bs), ShouldEqual, "something bad happened")
			})
		})

		Convey("isConnected", func() {
			Convey("Should return nil if transaction stream has not been initiated", func() {
				r := isConnected()
				So(r, ShouldEqual, false)
			})
		})
	})
}
