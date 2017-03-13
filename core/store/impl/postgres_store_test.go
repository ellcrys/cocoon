package impl

import (
	"testing"

	"github.com/ellcrys/util"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres" // gorm requires it
	"github.com/ncodes/cocoon/core/types/store"
	. "github.com/smartystreets/goconvey/convey"
)

func TestPosgresStore(t *testing.T) {
	Convey("PostgresStore", t, func() {

		var conStr = "host=localhost user=ned dbname=cocoon-dev sslmode=disable password="
		pgChain := new(PostgresStore)
		db, err := pgChain.Connect(conStr)
		So(err, ShouldBeNil)
		So(db, ShouldNotBeNil)

		var RestDB = func() {
			db.(*gorm.DB).DropTable(LedgerTableName, TransactionTableName)
		}

		Convey(".Connect", func() {
			Convey("should return error when unable to connect to a postgres server", func() {
				var conStr = "host=localhost user=wrong dbname=test sslmode=disable password=abc"
				pgChain := new(PostgresStore)
				db, err := pgChain.Connect(conStr)
				So(db, ShouldBeNil)
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "failed to connect to store backend")
			})
		})

		Convey(".Init", func() {

			Convey("when ledger table does not exists", func() {

				Convey("should create ledger and transactions table and create a global ledger entry", func() {

					ledgerEntryExists := db.(*gorm.DB).HasTable(LedgerTableName)
					So(ledgerEntryExists, ShouldEqual, false)

					err := pgChain.Init(store.GetGlobalLedgerName())
					So(err, ShouldBeNil)

					ledgerEntryExists = db.(*gorm.DB).HasTable(LedgerTableName)
					So(ledgerEntryExists, ShouldEqual, true)

					ledgerEntryExists = db.(*gorm.DB).HasTable(TransactionTableName)
					So(ledgerEntryExists, ShouldEqual, true)

					Convey("ledger table must include a global ledger entry", func() {
						var entries []store.Ledger
						err := db.(*gorm.DB).Find(&entries).Error
						So(err, ShouldBeNil)
						So(len(entries), ShouldEqual, 1)
						So(entries[0].Name, ShouldEqual, store.GetGlobalLedgerName())
					})

					Reset(func() {
						RestDB()
					})
				})
			})

			Convey("when ledger table exists", func() {
				Convey("should return nil with no effect", func() {
					err := pgChain.Init(store.GetGlobalLedgerName())
					So(err, ShouldBeNil)

					ledgerEntryExists := db.(*gorm.DB).HasTable(LedgerTableName)
					So(ledgerEntryExists, ShouldEqual, true)

					var entries []store.Ledger
					err = db.(*gorm.DB).Find(&entries).Error
					So(err, ShouldBeNil)
					So(len(entries), ShouldEqual, 1)
					So(entries[0].Name, ShouldEqual, store.GetGlobalLedgerName())
				})

				Reset(func() {
					RestDB()
				})
			})
		})

		Convey(".MakeLegderHash", func() {

			Convey("should return expected ledger hash", func() {
				hash := pgChain.MakeLegderHash(&store.Ledger{
					Name:      store.GetGlobalLedgerName(),
					Public:    true,
					CreatedAt: 1488196279,
				})
				So(hash, ShouldEqual, "2f97bb39bf93e995dcb611632fbb72424722b5e6090a0bc483846e46128eb74b")
			})

		})

		Convey(".CreateLedger", func() {

			var ledgerName = util.RandString(10)
			err := pgChain.Init(store.GetGlobalLedgerName())
			So(err, ShouldBeNil)

			Convey("should successfully create a ledger entry", func() {

				ledger, err := pgChain.CreateLedger(ledgerName, true)
				So(err, ShouldBeNil)
				So(ledger, ShouldNotBeNil)

				ledger, err = pgChain.CreateLedger(util.RandString(10), true)
				So(err, ShouldBeNil)
				So(ledger, ShouldNotBeNil)

				Convey("should return error since a ledger with same name already exists", func() {
					_, err := pgChain.CreateLedger(ledgerName, false)
					So(err, ShouldNotBeNil)
					So(err.Error(), ShouldEqual, `ledger with matching name already exists`)
				})
			})

			Reset(func() {
				RestDB()
			})
		})

		Convey(".GetLedger", func() {

			err := pgChain.Init(store.GetGlobalLedgerName())
			So(err, ShouldBeNil)

			Convey("should return nil when ledger does not exist", func() {
				tx, err := pgChain.GetLedger("wrong_name")
				So(tx, ShouldBeNil)
				So(err, ShouldBeNil)
			})

			Convey("should return expected transaction", func() {
				name := util.RandString(10)
				ledger, err := pgChain.CreateLedger(name, true)
				So(err, ShouldBeNil)
				So(ledger, ShouldNotBeNil)

				found, err := pgChain.GetLedger(name)
				So(found, ShouldNotBeNil)
				So(err, ShouldBeNil)
				So(found.Hash, ShouldEqual, ledger.Hash)
			})

			Reset(func() {
				RestDB()
			})
		})

		Convey(".MakeTxHash", func() {
			Convey("should return expected transaction hash", func() {
				hash := pgChain.MakeTxHash(&store.Transaction{
					ID:        util.Sha256("tx_id"),
					Key:       "balance",
					Value:     "30.50",
					CreatedAt: 12345678,
				})
				So(hash, ShouldEqual, "6d0318d08d26bd08bb0f90c2d86f9d6ffeebbc3a0767f1544cc76748d52670f5")
			})
		})

		Convey(".Put", func() {

			err := pgChain.Init(store.GetGlobalLedgerName())
			So(err, ShouldBeNil)

			Convey("expects new transaction to be the first and only transaction", func() {
				txID := util.Sha256("tx_id")
				ledger := util.Sha256("ledger_name")
				_, err := pgChain.Put(txID, ledger, "key", "value")
				So(err, ShouldBeNil)

				var allTx []store.Transaction
				err = db.(*gorm.DB).Find(&allTx).Error
				So(err, ShouldBeNil)
				So(len(allTx), ShouldEqual, 1)

				Convey("expects the only transaction to have expected fields set", func() {
					So(allTx[0].Hash, ShouldNotEqual, "")
					So(allTx[0].ID, ShouldEqual, txID)
					So(allTx[0].Ledger, ShouldEqual, ledger)
					So(allTx[0].Key, ShouldEqual, "key")
					So(allTx[0].Value, ShouldEqual, "value")
				})

			})

			Reset(func() {
				RestDB()
			})
		})

		Convey(".GetByID", func() {

			err := pgChain.Init(store.GetGlobalLedgerName())
			So(err, ShouldBeNil)

			Convey("should return nil when transaction does not exist", func() {
				tx, err := pgChain.GetByID("unknown_id")
				So(tx, ShouldBeNil)
				So(err, ShouldBeNil)
			})

			Convey("should return an expected transaction", func() {
				key := util.UUID4()
				txID := util.Sha256(util.UUID4())
				tx, err := pgChain.Put(txID, util.Sha256(util.RandString(5)), key, "value")
				So(tx, ShouldNotBeNil)
				So(err, ShouldBeNil)

				tx, err = pgChain.GetByID(txID)
				So(err, ShouldBeNil)
				So(tx, ShouldNotBeNil)
				So(tx.Key, ShouldEqual, key)
				So(tx.Value, ShouldEqual, "value")
			})

			Reset(func() {
				RestDB()
			})
		})

		Convey(".Get", func() {

			err := pgChain.Init(store.GetGlobalLedgerName())
			So(err, ShouldBeNil)

			Convey("should return nil when transaction does not exist", func() {
				tx, err := pgChain.Get(store.GetGlobalLedgerName(), "wrong_key")
				So(tx, ShouldBeNil)
				So(err, ShouldBeNil)
			})

			Convey("should return expected transaction", func() {
				key := util.UUID4()
				txID := util.Sha256(util.UUID4())
				ledger := util.Sha256(util.RandString(5))
				tx, err := pgChain.Put(txID, ledger, key, "value")
				So(tx, ShouldNotBeNil)
				So(err, ShouldBeNil)

				tx2, err := pgChain.Get(ledger, key)
				So(tx, ShouldNotBeNil)
				So(err, ShouldBeNil)
				So(tx2.Hash, ShouldEqual, tx.Hash)
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
