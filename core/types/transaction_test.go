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
				So(tx.MakeHash(), ShouldEqual, "25cea8016b574f5f68ca721976a53918cdd2d33724ea957c592413bf5ec4481a")
			})
		})
	})
}
