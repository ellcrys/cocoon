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
				So(tx.MakeHash(), ShouldEqual, "cb984a6f2f6ff2addd46e875b87550653f1345ad61892b2af2d4892b3d8b9e60")
			})
		})
	})
}
