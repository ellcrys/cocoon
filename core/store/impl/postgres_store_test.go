package impl

import (
	"database/sql"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/ellcrys/util"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres" // gorm requires it
	"github.com/ellcrys/cocoon/core/blockchain/impl"
	"github.com/ellcrys/cocoon/core/types"
	. "github.com/smartystreets/goconvey/convey"
)

var db *sql.DB
var dbName = "test_" + strings.ToLower(util.RandString(5))
var conStr = util.Env("STORE_CON_STR", "host=localhost user=ned sslmode=disable password=")
var conStrWithDB = util.Env("STORE_CON_STR", "host=localhost database="+dbName+" user=ned sslmode=disable password=")

func createDb(t *testing.T) error {
	_, err := db.Query(fmt.Sprintf("CREATE DATABASE %s;", dbName))
	return err
}

func dropDB(t *testing.T) error {
	_, err := db.Query(fmt.Sprintf("DROP DATABASE %s;", dbName))
	return err
}

func init() {
	os.Setenv("ENV", "test")

	var err error
	db, err = sql.Open("postgres", conStr)
	if err != nil {
		panic(fmt.Errorf("failed to connect to database: %s", err))
	}
}

func TestPostgresStore(t *testing.T) {

	defer db.Close()
	err := createDb(t)
	if err != nil {
		t.Fatal(err)
	}
	defer dropDB(t)

	pgStore := new(PostgresStore)
	db, err := pgStore.Connect(conStrWithDB)
	if err != nil {
		t.Fatal("failed to connect to pg store")
	}

	pgChain := new(impl.PostgresBlockchain)
	_, err = pgChain.Connect(conStrWithDB)
	if err != nil {
		t.Fatal("failed to connect to pg blockchain")
	}

	err = pgChain.Init()
	if err != nil {
		t.Fatal("failed to initialize pg blockchain")
	}

	Convey("PostgresStore", t, func() {

		pgStore.SetBlockchainImplementation(pgChain)

		Convey(".Init", func() {
			err := pgStore.Init(types.GetSystemPublicLedgerName(), types.GetSystemPrivateLedgerName())
			So(err, ShouldBeNil)
			var entries []types.Ledger
			err = db.(*gorm.DB).Find(&entries).Error
			So(err, ShouldBeNil)
			So(len(entries), ShouldEqual, 2)
			So(entries[0].Name, ShouldEqual, types.GetSystemPublicLedgerName())
			So(entries[1].Name, ShouldEqual, types.GetSystemPrivateLedgerName())
		})

		Convey(".MakeTxKey", func() {
			Convey("should create expected tx key", func() {
				key := types.MakeTxKey("namespace", "accounts")
				So(key, ShouldEqual, "namespace;accounts")
			})
		})

		Convey(".GetActualKeyFromTxKey", func() {
			Convey("should return expected key", func() {
				txKey := types.MakeTxKey("namespace", "accounts")
				key := types.GetActualKeyFromTxKey(txKey)
				So(key, ShouldEqual, "accounts")
			})
		})

		Convey(".CreateLedger", func() {

			var ledgerName = util.RandString(10)

			Convey("should successfully create a ledger", func() {
				ledger, err := pgStore.CreateLedger(util.RandString(10), ledgerName, true, true)
				So(err, ShouldBeNil)
				So(ledger, ShouldNotBeNil)
				So(ledger.Chained, ShouldEqual, true)
				So(ledger.Public, ShouldEqual, true)

				ledger, err = pgStore.CreateLedger(util.RandString(10), util.RandString(10), true, true)
				So(err, ShouldBeNil)
				So(ledger, ShouldNotBeNil)

				Convey("should return error since a ledger with same name already exists", func() {
					_, err := pgStore.CreateLedger(util.RandString(10), ledgerName, false, false)
					So(err, ShouldNotBeNil)
					So(err.Error(), ShouldEqual, `ledger with matching name already exists`)
				})
			})
		})

		Convey(".CreateLedgerThen", func() {
			var ledgerName = util.RandString(10)

			Convey("should fail to create a ledger if thenFunction returns an error", func() {
				var ErrFromThenFunc = fmt.Errorf("thenFunc error")
				ledger, err := pgStore.CreateLedgerThen("cocoon_id", ledgerName, true, true, func() error {
					return ErrFromThenFunc
				})
				So(err, ShouldNotBeNil)
				So(err, ShouldResemble, ErrFromThenFunc)
				So(ledger, ShouldBeNil)
			})

			Convey("should successfully create a ledger if then function does not return error", func() {
				ledger, err := pgStore.CreateLedgerThen(util.RandString(5), ledgerName, true, true, func() error {
					return nil
				})
				So(err, ShouldBeNil)
				So(ledger, ShouldNotBeNil)
				So(ledger.Chained, ShouldEqual, true)
				So(ledger.Public, ShouldEqual, true)
			})
		})

		Convey(".GetLedger", func() {

			err := pgStore.Init(types.GetSystemPublicLedgerName(), types.GetSystemPrivateLedgerName())
			So(err, ShouldBeNil)

			Convey("should return nil when ledger does not exist", func() {
				tx, err := pgStore.GetLedger("wrong_name")
				So(tx, ShouldBeNil)
				So(err, ShouldBeNil)
			})

			Convey("should return existing ledger", func() {
				name := util.RandString(10)
				ledger, err := pgStore.CreateLedger("abc", name, true, true)
				So(err, ShouldBeNil)
				So(ledger, ShouldNotBeNil)

				found, err := pgStore.GetLedger(name)
				So(found, ShouldNotBeNil)
				So(err, ShouldBeNil)
				So(found.Number, ShouldEqual, ledger.Number)
				So(found.CocoonID, ShouldEqual, ledger.CocoonID)
			})
		})

		Convey(".Put", func() {

			err := pgStore.Init(types.GetSystemPublicLedgerName(), types.GetSystemPrivateLedgerName())
			So(err, ShouldBeNil)

			Convey("expects new transaction to be the first and only transaction", func() {
				ledger := util.Sha256("ledger_name")
				_, err := pgStore.CreateLedger(util.RandString(5), ledger, true, true)
				So(err, ShouldBeNil)

				tx := &types.Transaction{ID: util.Sha256(util.UUID4()), Key: "key", Value: "value"}
				txs, err := pgStore.Put(ledger, []*types.Transaction{tx})
				So(err, ShouldBeNil)
				So(len(txs), ShouldEqual, 1)

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
		})

		Convey(".PutThen", func() {

			err := pgStore.Init(types.GetSystemPublicLedgerName(), types.GetSystemPrivateLedgerName())
			So(err, ShouldBeNil)

			ledger := util.RandString(5)
			_, err = pgStore.CreateLedger(util.RandString(5), ledger, true, true)
			So(err, ShouldBeNil)

			Convey("Should fail if thenFunc returns error", func() {
				var ErrThenFunc = fmt.Errorf("thenFunc error")
				tx := &types.Transaction{ID: util.Sha256(util.UUID4()), Key: "key", Value: "value"}
				_, err := pgStore.PutThen(ledger, []*types.Transaction{tx}, func([]*types.Transaction) error {
					return ErrThenFunc
				})
				So(err, ShouldResemble, ErrThenFunc)
			})

			Convey("Should successfully add transaction if thenFunc does not return error", func() {
				tx := &types.Transaction{ID: util.Sha256(util.UUID4()), Key: "key", Value: "value"}
				txs, err := pgStore.PutThen(ledger, []*types.Transaction{tx}, func([]*types.Transaction) error {
					return nil
				})
				So(err, ShouldBeNil)
				So(len(txs), ShouldEqual, 1)
				So(txs[0].Err, ShouldBeEmpty)
				So(txs[0].ID, ShouldEqual, tx.ID)
			})
		})

		Convey(".Get", func() {

			err := pgStore.Init(types.GetSystemPublicLedgerName(), types.GetSystemPrivateLedgerName())
			So(err, ShouldBeNil)

			Convey("should return nil when transaction does not exist", func() {
				tx, err := pgStore.Get(types.GetSystemPublicLedgerName(), "wrong_key")
				So(tx, ShouldBeNil)
				So(err, ShouldBeNil)
			})

			Convey("should return expected transaction", func() {
				key := util.UUID4()

				ledger := util.Sha256(util.RandString(5))
				_, err = pgStore.CreateLedger(util.RandString(5), ledger, true, true)
				So(err, ShouldBeNil)

				tx := &types.Transaction{ID: util.Sha256(util.UUID4()), Key: key, Value: "value"}
				txs, err := pgStore.Put(ledger, []*types.Transaction{tx})
				So(err, ShouldBeNil)
				So(len(txs), ShouldEqual, 1)
				So(txs[0].Err, ShouldBeEmpty)
				So(txs[0].ID, ShouldEqual, tx.ID)

				tx2, err := pgStore.Get(ledger, key)
				So(tx, ShouldNotBeNil)
				So(err, ShouldBeNil)
				So(tx2.Hash, ShouldEqual, tx.Hash)
			})
		})

		Convey(".GetRange", func() {

			err := pgStore.Init(types.GetSystemPublicLedgerName(), types.GetSystemPrivateLedgerName())
			So(err, ShouldBeNil)

			Convey("Should successfully return expected transactions and exclude end key when `includeEndKey` is false", func() {
				ledger := util.Sha256(util.UUID4())
				tx := &types.Transaction{ID: util.Sha256(util.UUID4()), Key: "account.ken", Value: "100"}
				tx2 := &types.Transaction{ID: util.Sha256(util.UUID4()), Key: "account.ben", Value: "110"}
				tx3 := &types.Transaction{ID: util.Sha256(util.UUID4()), Key: "account.glen", Value: "200"}
				tx4 := &types.Transaction{ID: util.Sha256(util.UUID4()), Key: "z", Value: "200"}
				_, err := pgStore.Put(ledger, []*types.Transaction{tx, tx2, tx3, tx4})
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
				_, err := pgStore.Put(ledger, []*types.Transaction{tx, tx2, tx3, tx4})
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
				_, err := pgStore.Put(ledger, []*types.Transaction{tx, tx2, tx3, tx4})
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
				_, err := pgStore.Put(ledger, []*types.Transaction{tx, tx2, tx3})
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
				_, err := pgStore.Put(ledger, []*types.Transaction{tx, tx2, tx3})
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
		})
	})
}
