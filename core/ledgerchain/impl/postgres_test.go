package impl

import (
	"testing"

	"github.com/ellcrys/util"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres" // gorm requires it
	. "github.com/smartystreets/goconvey/convey"
)

func TestPosgresLedgerChain(t *testing.T) {
	Convey("PostgresLedgerChain", t, func() {

		var conStr = "host=localhost user=ned dbname=cocoon-dev sslmode=disable password="
		pgChain := new(PostgresLedgerChain)
		db, err := pgChain.Connect(conStr)
		So(err, ShouldBeNil)
		So(db, ShouldNotBeNil)

		var RestDB = func() {
			db.(*gorm.DB).DropTable(LedgerTableName, TransactionTableName)
		}

		Convey(".Connect", func() {
			Convey("should return error when unable to connect to a postgres server", func() {
				var conStr = "host=localhost user=wrong dbname=test sslmode=disable password=abc"
				pgChain := new(PostgresLedgerChain)
				db, err := pgChain.Connect(conStr)
				So(db, ShouldBeNil)
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "failed to connect to ledgerchain backend")
			})
		})

		Convey(".Init", func() {

			Convey("when ledger table does not exists", func() {

				Convey("should create ledger and transactions table and create a global ledger entry", func() {

					ledgerEntryExists := db.(*gorm.DB).HasTable(LedgerTableName)
					So(ledgerEntryExists, ShouldEqual, false)

					err := pgChain.Init()
					So(err, ShouldBeNil)

					ledgerEntryExists = db.(*gorm.DB).HasTable(LedgerTableName)
					So(ledgerEntryExists, ShouldEqual, true)

					ledgerEntryExists = db.(*gorm.DB).HasTable(TransactionTableName)
					So(ledgerEntryExists, ShouldEqual, true)

					Convey("ledger table must include a global ledger entry", func() {
						var entries []Ledger
						err := db.(*gorm.DB).Find(&entries).Error
						So(err, ShouldBeNil)
						So(len(entries), ShouldEqual, 1)
						So(entries[0].Name, ShouldEqual, "global")
					})

					Reset(func() {
						RestDB()
					})
				})
			})

			Convey("when ledger table exists", func() {
				Convey("should return nil with no effect", func() {
					err := pgChain.Init()
					So(err, ShouldBeNil)

					ledgerEntryExists := db.(*gorm.DB).HasTable(LedgerTableName)
					So(ledgerEntryExists, ShouldEqual, true)

					var entries []Ledger
					err = db.(*gorm.DB).Find(&entries).Error
					So(err, ShouldBeNil)
					So(len(entries), ShouldEqual, 1)
					So(entries[0].Name, ShouldEqual, "global")
				})

				Reset(func() {
					RestDB()
				})
			})
		})

		Convey(".MakeLegderHash", func() {

			Convey("should return expected ledger hash", func() {
				hash := pgChain.MakeLegderHash(&Ledger{
					PrevLedgerHash: NullHash,
					Name:           "global",
					CocoonCodeID:   "xh6549dh",
					Public:         true,
					CreatedAt:      1488196279,
				})
				So(hash, ShouldEqual, "8741cc99b2d5cae9c49d05930cf014b87e60d20fecc122deb0ff3beaee7c8064")
			})

		})

		Convey(".CreateLedger", func() {

			var ledgerName = util.RandString(10)
			err := pgChain.Init()
			So(err, ShouldBeNil)

			Convey("should successfully create a ledger entry", func() {

				ledger, err := pgChain.CreateLedger(ledgerName, util.RandString(10), true)
				So(err, ShouldBeNil)
				So(ledger, ShouldNotBeNil)

				ledger, err = pgChain.CreateLedger(util.RandString(10), util.RandString(10), true)
				So(err, ShouldBeNil)
				So(ledger, ShouldNotBeNil)

				Convey("should return error since a ledger with same name already exists", func() {
					_, err := pgChain.CreateLedger(ledgerName, util.RandString(10), true)
					So(err, ShouldNotBeNil)
					So(err.Error(), ShouldEqual, `pq: duplicate key value violates unique constraint "uix_ledgers_name"`)
				})
			})

			Convey("should successfully add a new ledger entry but PrevLedgerHash column must reference the previous ledger", func() {
				ledger, err := pgChain.CreateLedger(util.RandString(10), util.RandString(10), true)
				So(err, ShouldBeNil)
				So(ledger, ShouldNotBeNil)

				ledger2, err := pgChain.CreateLedger(util.RandString(10), util.RandString(10), true)
				So(err, ShouldBeNil)
				So(ledger2, ShouldNotBeNil)
				So(ledger2.(*Ledger).PrevLedgerHash, ShouldEqual, ledger.(*Ledger).Hash)
			})

			Reset(func() {
				RestDB()
			})
		})

		Convey(".Put", func() {

			err := pgChain.Init()
			So(err, ShouldBeNil)

			Convey("should return error if ledger does not exists", func() {
				err := pgChain.Put("unknown", "key", "value")
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "ledger does not exist")
			})
		})

		Convey(".MakeTxHash", func() {

			Convey("should return expected transaction hash", func() {
				hash := pgChain.MakeTxHash(&Transaction{
					Ledger:     "global",
					ID:         util.Sha256("tx_id"),
					Key:        "balance",
					Value:      "30.50",
					PrevTxHash: util.Sha256("prev_tx_hash"),
				})
				So(hash, ShouldEqual, "b472517c05fdc297c10edbf2cb359f9d85efb1879507a8b3d48ca93df0f462af")
			})

		})

		Reset(func() {
			RestDB()
		})
	})
}
