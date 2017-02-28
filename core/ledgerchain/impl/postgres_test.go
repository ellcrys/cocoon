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

		Convey(".GetLedger", func() {

			err := pgChain.Init()
			So(err, ShouldBeNil)

			Convey("should return nil when ledger does not exist", func() {
				tx, err := pgChain.GetLedger("wrong_name")
				So(tx, ShouldBeNil)
				So(err, ShouldBeNil)
			})

			Convey("should return expected transaction", func() {
				name := util.RandString(10)
				cocoonCodeID := util.RandString(10)
				ledger, err := pgChain.CreateLedger(name, cocoonCodeID, true)
				So(err, ShouldBeNil)
				So(ledger, ShouldNotBeNil)

				found, err := pgChain.GetLedger(name)
				So(found, ShouldNotBeNil)
				So(err, ShouldBeNil)
				So(found.(*Ledger).Hash, ShouldEqual, ledger.(*Ledger).Hash)
			})

			Reset(func() {
				RestDB()
			})
		})

		Convey(".MakeTxHash", func() {
			Convey("should return expected transaction hash", func() {
				hash := pgChain.MakeTxHash(&Transaction{
					ID:         util.Sha256("tx_id"),
					Key:        "balance",
					Value:      "30.50",
					PrevTxHash: util.Sha256("prev_tx_hash"),
					CreatedAt:  12345678,
				})
				So(hash, ShouldEqual, "0e92cf41a17865093e4d4e6553f00da9dc66e8ae2d30d19ad36dc21a2e3c2b89")
			})
		})

		Convey(".Put", func() {

			err := pgChain.Init()
			So(err, ShouldBeNil)

			Convey("expects new transaction to be the first and only transaction", func() {
				txID := util.Sha256("tx_id")
				_, err := pgChain.Put(txID, "key", "value")
				So(err, ShouldBeNil)

				var allTx []Transaction
				err = db.(*gorm.DB).Find(&allTx).Error
				So(err, ShouldBeNil)
				So(len(allTx), ShouldEqual, 1)

				Convey("expects the only transaction to have expected fields set", func() {
					So(allTx[0].Hash, ShouldNotEqual, "")
					So(allTx[0].ID, ShouldEqual, txID)
					So(allTx[0].Key, ShouldEqual, "key")
					So(allTx[0].Value, ShouldEqual, "value")
					So(allTx[0].NextTxHash, ShouldEqual, "")
					So(allTx[0].PrevTxHash, ShouldEqual, NullHash)
				})

				Convey("expects new transaction to have its PrevTxHash set to the hash of the last transaction's hash", func() {
					tx, err := pgChain.Put(util.Sha256(util.RandString(2)), "key", "value")
					So(err, ShouldBeNil)
					So(tx.(*Transaction).PrevTxHash, ShouldEqual, allTx[0].Hash)
				})
			})

			Reset(func() {
				RestDB()
			})
		})

		Convey(".GetByID", func() {

			err := pgChain.Init()
			So(err, ShouldBeNil)

			Convey("should return nil when transaction does not exist", func() {
				tx, err := pgChain.GetByID("unknown_id")
				So(tx, ShouldBeNil)
				So(err, ShouldBeNil)
			})

			Convey("should return an expected transaction", func() {
				key := util.UUID4()
				txID := util.Sha256(util.UUID4())
				tx, err := pgChain.Put(txID, key, "value")
				So(tx, ShouldNotBeNil)
				So(err, ShouldBeNil)

				tx, err = pgChain.GetByID(txID)
				So(err, ShouldBeNil)
				So(tx, ShouldNotBeNil)
				So(tx.(*Transaction).Key, ShouldEqual, key)
				So(tx.(*Transaction).Value, ShouldEqual, "value")
			})

			Reset(func() {
				RestDB()
			})
		})

		Convey(".Get", func() {

			err := pgChain.Init()
			So(err, ShouldBeNil)

			Convey("should return nil when transaction does not exist", func() {
				tx, err := pgChain.Get("wrong_key")
				So(tx, ShouldBeNil)
				So(err, ShouldBeNil)
			})

			Convey("should return expected transaction", func() {
				key := util.UUID4()
				txID := util.Sha256(util.UUID4())
				tx, err := pgChain.Put(txID, key, "value")
				So(tx, ShouldNotBeNil)
				So(err, ShouldBeNil)

				tx2, err := pgChain.Get(key)
				So(tx, ShouldNotBeNil)
				So(err, ShouldBeNil)
				So(tx2.(*Transaction).Hash, ShouldEqual, tx.(*Transaction).Hash)
			})

			Reset(func() {
				RestDB()
			})
		})

		Reset(func() {
			RestDB()
		})
	})
}
