package golang

import (
	"testing"

	"github.com/ncodes/cocoon/core/types/txchain"
	. "github.com/smartystreets/goconvey/convey"
)

func TestStub(t *testing.T) {
	Convey("GoStub", t, func() {

		Convey("GetGlobalLedgerName", func() {
			Convey("Should return the expected value set in txchain.GetGlobalLedgerName()", func() {
				So(GetGlobalLedgerName(), ShouldEqual, txchain.GetGlobalLedgerName())
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
