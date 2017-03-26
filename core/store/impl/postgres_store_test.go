package impl

import (
	"testing"

	"fmt"

	"github.com/ellcrys/util"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres" // gorm requires it
	"github.com/ncodes/cocoon/core/types"
	. "github.com/smartystreets/goconvey/convey"
)

func TestPosgresStore(t *testing.T) {
	Convey("PostgresStore", t, func() {

		var conStr = "host=localhost user=ned dbname=cocoon-dev sslmode=disable password="
		pgStore := new(PostgresStore)
		db, err := pgStore.Connect(conStr)
		So(err, ShouldBeNil)
		So(db, ShouldNotBeNil)

		var RestDB = func() {
			db.(*gorm.DB).DropTable(LedgerTableName, TransactionTableName)
		}

		Convey(".Connect", func() {
			Convey("should return error when unable to connect to a postgres server", func() {
				var conStr = "host=localhost user=wrong dbname=test sslmode=disable password=abc"
				pgStore := new(PostgresStore)
				db, err := pgStore.Connect(conStr)
				So(db, ShouldBeNil)
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, `failed to connect to store backend. pq: role "wrong" does not exist`)
			})
		})

		Convey(".Init", func() {

			Convey("when ledger table does not exists", func() {

				Convey("should create ledger and transactions table and create a global ledger entry", func() {

					ledgerEntryExists := db.(*gorm.DB).HasTable(LedgerTableName)
					So(ledgerEntryExists, ShouldEqual, false)

					err := pgStore.Init(types.GetGlobalLedgerName())
					So(err, ShouldBeNil)

					ledgerEntryExists = db.(*gorm.DB).HasTable(LedgerTableName)
					So(ledgerEntryExists, ShouldEqual, true)

					ledgerEntryExists = db.(*gorm.DB).HasTable(TransactionTableName)
					So(ledgerEntryExists, ShouldEqual, true)

					Convey("ledger table must include a global ledger entry", func() {
						var entries []types.Ledger
						err := db.(*gorm.DB).Find(&entries).Error
						So(err, ShouldBeNil)
						So(len(entries), ShouldEqual, 1)
						So(entries[0].Name, ShouldEqual, types.GetGlobalLedgerName())
					})

					Reset(func() {
						RestDB()
					})
				})
			})

			Convey("when ledger table exists", func() {
				Convey("should return nil with no effect", func() {
					err := pgStore.Init(types.GetGlobalLedgerName())
					So(err, ShouldBeNil)

					ledgerEntryExists := db.(*gorm.DB).HasTable(LedgerTableName)
					So(ledgerEntryExists, ShouldEqual, true)

					var entries []types.Ledger
					err = db.(*gorm.DB).Find(&entries).Error
					So(err, ShouldBeNil)
					So(len(entries), ShouldEqual, 1)
					So(entries[0].Name, ShouldEqual, types.GetGlobalLedgerName())
				})

				Reset(func() {
					RestDB()
				})
			})
		})

		Convey(".MakeLegderHash", func() {

			Convey("should return expected ledger hash", func() {
				hash := pgStore.MakeLegderHash(&types.Ledger{
					Name:      types.GetGlobalLedgerName(),
					Public:    true,
					CreatedAt: 1488196279,
				})
				So(hash, ShouldEqual, "2f97bb39bf93e995dcb611632fbb72424722b5e6090a0bc483846e46128eb74b")
			})

		})

		Convey(".MakeTxKey", func() {
			Convey("should create expected tx key", func() {
				key := pgStore.MakeTxKey("namespace", "accounts")
				So(key, ShouldEqual, "namespace.accounts")
			})
		})

		Convey(".GetActualKeyFromTxKey", func() {
			Convey("should return expected key", func() {
				txKey := pgStore.MakeTxKey("namespace", "accounts")
				key := pgStore.GetActualKeyFromTxKey(txKey)
				So(key, ShouldEqual, "accounts")
			})
		})

		Convey(".CreateLedger", func() {

			var ledgerName = util.RandString(10)
			err := pgStore.Init(types.GetGlobalLedgerName())
			So(err, ShouldBeNil)

			Convey("should successfully create a ledger entry", func() {

				ledger, err := pgStore.CreateLedger(ledgerName, true, true)
				So(err, ShouldBeNil)
				So(ledger, ShouldNotBeNil)
				So(ledger.Chained, ShouldEqual, true)
				So(ledger.Public, ShouldEqual, true)

				ledger, err = pgStore.CreateLedger(util.RandString(10), true, true)
				So(err, ShouldBeNil)
				So(ledger, ShouldNotBeNil)

				Convey("should return error since a ledger with same name already exists", func() {
					_, err := pgStore.CreateLedger(ledgerName, false, false)
					So(err, ShouldNotBeNil)
					So(err.Error(), ShouldEqual, `ledger with matching name already exists`)
				})
			})

			Reset(func() {
				RestDB()
			})
		})

		Convey(".CreateLedgerThen", func() {

			var ledgerName = util.RandString(10)
			err := pgStore.Init(types.GetGlobalLedgerName())
			So(err, ShouldBeNil)

			Convey("should fail to create a ledger if thenFunction returns an error", func() {
				var ErrFromThenFunc = fmt.Errorf("thenFunc error")
				ledger, err := pgStore.CreateLedgerThen(ledgerName, true, true, func() error {
					return ErrFromThenFunc
				})
				So(err, ShouldNotBeNil)
				So(err, ShouldResemble, ErrFromThenFunc)
				So(ledger, ShouldBeNil)
			})

			Convey("should successfully create a ledger if then function does not return error", func() {
				ledger, err := pgStore.CreateLedgerThen(ledgerName, true, true, func() error {
					return nil
				})
				So(err, ShouldBeNil)
				So(ledger, ShouldNotBeNil)
				So(ledger.Chained, ShouldEqual, true)
				So(ledger.Public, ShouldEqual, true)
			})

			Reset(func() {
				RestDB()
			})
		})

		Convey(".GetLedger", func() {

			err := pgStore.Init(types.GetGlobalLedgerName())
			So(err, ShouldBeNil)

			Convey("should return nil when ledger does not exist", func() {
				tx, err := pgStore.GetLedger("wrong_name")
				So(tx, ShouldBeNil)
				So(err, ShouldBeNil)
			})

			Convey("should return existing ledger", func() {
				name := util.RandString(10)
				ledger, err := pgStore.CreateLedger(name, true, true)
				So(err, ShouldBeNil)
				So(ledger, ShouldNotBeNil)

				found, err := pgStore.GetLedger(name)
				So(found, ShouldNotBeNil)
				So(err, ShouldBeNil)
				So(found.Hash, ShouldEqual, ledger.Hash)
			})

			Reset(func() {
				RestDB()
			})
		})

		Convey(".Put", func() {

			err := pgStore.Init(types.GetGlobalLedgerName())
			So(err, ShouldBeNil)

			Convey("expects new transaction to be the first and only transaction", func() {
				ledger := util.Sha256("ledger_name")
				_, err := pgStore.CreateLedger(ledger, true, true)
				So(err, ShouldBeNil)

				tx := &types.Transaction{ID: util.Sha256(util.UUID4()), Key: "key", Value: "value"}
				err = pgStore.Put(ledger, []*types.Transaction{tx})
				So(err, ShouldBeNil)

				var allTx []types.Transaction
				err = db.(*gorm.DB).Find(&allTx).Error
				So(err, ShouldBeNil)
				So(len(allTx), ShouldEqual, 1)

				Convey("expects the only transaction to have expected fields set", func() {
					So(allTx[0].Hash, ShouldNotEqual, "")
					So(allTx[0].ID, ShouldEqual, tx.ID)
					So(allTx[0].Ledger, ShouldEqual, ledger)
					So(allTx[0].Key, ShouldEqual, "key")
					So(allTx[0].Value, ShouldEqual, "value")
				})

			})

			Reset(func() {
				RestDB()
			})
		})

		Convey(".PutThen", func() {

			err := pgStore.Init(types.GetGlobalLedgerName())
			So(err, ShouldBeNil)

			ledger := util.Sha256("ledger_name")
			_, err = pgStore.CreateLedger(ledger, true, true)
			So(err, ShouldBeNil)

			Convey("Should fail if thenFunc returns error", func() {
				var ErrThenFunc = fmt.Errorf("thenFunc error")
				tx := &types.Transaction{ID: util.Sha256(util.UUID4()), Key: "key", Value: "value"}
				err = pgStore.PutThen(ledger, []*types.Transaction{tx}, func() error {
					return ErrThenFunc
				})
				So(err, ShouldResemble, ErrThenFunc)
			})

			Convey("Should successfully add transaction if thenFunc does not return error", func() {
				tx := &types.Transaction{ID: util.Sha256(util.UUID4()), Key: "key", Value: "value"}
				err = pgStore.PutThen(ledger, []*types.Transaction{tx}, func() error {
					return nil
				})
				So(err, ShouldBeNil)

				var allTx []types.Transaction
				err = db.(*gorm.DB).Find(&allTx).Error
				So(err, ShouldBeNil)
				So(len(allTx), ShouldEqual, 1)
			})

			Reset(func() {
				RestDB()
			})
		})

		Convey(".GetByID", func() {

			err := pgStore.Init(types.GetGlobalLedgerName())
			So(err, ShouldBeNil)

			Convey("should return nil when transaction does not exist", func() {
				tx, err := pgStore.GetByID(types.GetGlobalLedgerName(), "unknown_id")
				So(tx, ShouldBeNil)
				So(err, ShouldBeNil)
			})

			Convey("should return an expected transaction", func() {
				txID := util.Sha256(util.UUID4())
				tx := &types.Transaction{ID: txID, Key: "key", Value: "value"}

				ledger := util.Sha256(util.RandString(5))
				_, err := pgStore.CreateLedger(ledger, true, true)
				So(err, ShouldBeNil)

				err = pgStore.Put(ledger, []*types.Transaction{tx})
				So(err, ShouldBeNil)

				tx2, err := pgStore.GetByID(ledger, txID)
				So(err, ShouldBeNil)
				So(tx2, ShouldNotBeNil)
				So(tx2.Key, ShouldEqual, tx.Key)
				So(tx2.Value, ShouldEqual, tx.Value)
				So(tx2.Hash, ShouldNotBeEmpty)
			})

			Reset(func() {
				RestDB()
			})
		})

		Convey(".Get", func() {

			err := pgStore.Init(types.GetGlobalLedgerName())
			So(err, ShouldBeNil)

			Convey("should return nil when transaction does not exist", func() {
				tx, err := pgStore.Get(types.GetGlobalLedgerName(), "wrong_key")
				So(tx, ShouldBeNil)
				So(err, ShouldBeNil)
			})

			Convey("should return expected transaction", func() {
				key := util.UUID4()

				ledger := util.Sha256(util.RandString(5))
				_, err = pgStore.CreateLedger(ledger, true, true)
				So(err, ShouldBeNil)

				tx := &types.Transaction{ID: util.Sha256(util.UUID4()), Key: key, Value: "value"}
				err := pgStore.Put(ledger, []*types.Transaction{tx})
				So(tx, ShouldNotBeNil)
				So(err, ShouldBeNil)

				tx2, err := pgStore.Get(ledger, key)
				So(tx, ShouldNotBeNil)
				So(err, ShouldBeNil)
				So(tx2.Hash, ShouldEqual, tx.Hash)
			})

			Reset(func() {
				RestDB()
			})
		})

		Convey(".GetRange", func() {

			err := pgStore.Init(types.GetGlobalLedgerName())
			So(err, ShouldBeNil)

			Convey("Should successfully return expected transactions and exclude end key when `includeEndKey` is false", func() {
				ledger := util.Sha256(util.UUID4())
				tx := &types.Transaction{ID: util.Sha256(util.UUID4()), Key: "account.ken", Value: "100"}
				tx2 := &types.Transaction{ID: util.Sha256(util.UUID4()), Key: "account.ben", Value: "110"}
				tx3 := &types.Transaction{ID: util.Sha256(util.UUID4()), Key: "account.glen", Value: "200"}
				tx4 := &types.Transaction{ID: util.Sha256(util.UUID4()), Key: "z", Value: "200"}
				err := pgStore.Put(ledger, []*types.Transaction{tx, tx2, tx3, tx4})
				So(tx, ShouldNotBeNil)
				So(err, ShouldBeNil)

				txs, err := pgStore.GetRange(ledger, "a", "z", false, 10, 0)
				So(err, ShouldBeNil)
				So(len(txs), ShouldEqual, 3)
			})

			Convey("Should successfully return expected transactions and include end key when `includeEndKey` is true", func() {
				ledger := util.Sha256(util.UUID4())
				tx := &types.Transaction{ID: util.Sha256(util.UUID4()), Key: "account.ken", Value: "100"}
				tx2 := &types.Transaction{ID: util.Sha256(util.UUID4()), Key: "account.ben", Value: "110"}
				tx3 := &types.Transaction{ID: util.Sha256(util.UUID4()), Key: "account.glen", Value: "200"}
				tx4 := &types.Transaction{ID: util.Sha256(util.UUID4()), Key: "z", Value: "200"}
				err := pgStore.Put(ledger, []*types.Transaction{tx, tx2, tx3, tx4})
				So(tx, ShouldNotBeNil)
				So(err, ShouldBeNil)

				txs, err := pgStore.GetRange(ledger, "account", "z", true, 10, 0)
				So(err, ShouldBeNil)
				So(len(txs), ShouldEqual, 4)
			})

			Convey("Should successfully return transactions with matching startKey if only startKey is provided` is true", func() {
				ledger := util.Sha256(util.UUID4())
				tx := &types.Transaction{ID: util.Sha256(util.UUID4()), Key: "account.ken", Value: "100"}
				tx2 := &types.Transaction{ID: util.Sha256(util.UUID4()), Key: "account.ben", Value: "110"}
				tx3 := &types.Transaction{ID: util.Sha256(util.UUID4()), Key: "account.glen", Value: "200"}
				tx4 := &types.Transaction{ID: util.Sha256(util.UUID4()), Key: "z", Value: "200"}
				err := pgStore.Put(ledger, []*types.Transaction{tx, tx2, tx3, tx4})
				So(tx, ShouldNotBeNil)
				So(err, ShouldBeNil)

				txs, err := pgStore.GetRange(ledger, "account", "", true, 10, 0)
				So(err, ShouldBeNil)
				So(len(txs), ShouldEqual, 3)
			})

			Convey("Should successfully return transactions with matching endKey if only endKey is provided` is true", func() {
				ledger := util.Sha256(util.UUID4())
				tx := &types.Transaction{ID: util.Sha256(util.UUID4()), Key: "account.ken", Value: "100"}
				tx2 := &types.Transaction{ID: util.Sha256(util.UUID4()), Key: "ben.account", Value: "110"}
				tx3 := &types.Transaction{ID: util.Sha256(util.UUID4()), Key: "glen.account", Value: "200"}
				err := pgStore.Put(ledger, []*types.Transaction{tx, tx2, tx3})
				So(tx, ShouldNotBeNil)
				So(err, ShouldBeNil)

				txs, err := pgStore.GetRange(ledger, "", "%account", true, 10, 0)
				So(err, ShouldBeNil)
				So(len(txs), ShouldEqual, 2)
			})

			Convey("Should successfully return expected transaction limit ", func() {
				ledger := util.Sha256(util.UUID4())
				tx := &types.Transaction{ID: util.Sha256(util.UUID4()), Key: "account.ken", Value: "100"}
				tx2 := &types.Transaction{ID: util.Sha256(util.UUID4()), Key: "ben.account", Value: "110"}
				tx3 := &types.Transaction{ID: util.Sha256(util.UUID4()), Key: "glen.account", Value: "200"}
				err := pgStore.Put(ledger, []*types.Transaction{tx, tx2, tx3})
				So(tx, ShouldNotBeNil)
				So(err, ShouldBeNil)

				txs, err := pgStore.GetRange(ledger, "", "%account", true, 1, 0)
				So(err, ShouldBeNil)
				So(len(txs), ShouldEqual, 1)
				So(txs[0].Key, ShouldEqual, "ben.account")

				txs, err = pgStore.GetRange(ledger, "", "%account", true, 1, 1)
				So(err, ShouldBeNil)
				So(len(txs), ShouldEqual, 1)
				So(txs[0].Key, ShouldEqual, "glen.account")
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
