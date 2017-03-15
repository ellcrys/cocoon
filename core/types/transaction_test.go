package types

import (
	"testing"

	_ "github.com/jinzhu/gorm/dialects/postgres" // gorm requires it
	. "github.com/smartystreets/goconvey/convey"
)

func TestTransaction(t *testing.T) {
	Convey("Transaction", t, func() {
		Convey(".MakeHash", func() {
			Convey("Should successfully create a hash", func() {
				tx := &Transaction{
					Number:    1,
					Ledger:    "ledger1",
					ID:        "some_id",
					Key:       "key",
					Value:     "value",
					CreatedAt: 123456789,
				}
				So(tx.MakeHash(), ShouldEqual, "c29acdab4705ef20e6afd593eaec4c631a1d07f211b0ff024d0185da2909b825")
			})
		})
	})
}
