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

func TestPosgresBlockchain(t *testing.T) {

	defer db.Close()
	err := createDb(t)
	if err != nil {
		t.Fatal(err)
	}
	defer dropDB(t)

	pgChain := new(PostgresBlockchain)
	db, err := pgChain.Connect(conStrWithDB)
	if err != nil {
		t.Fatal("failed to connect to pg blockchain")
	}

	Convey("PostgresBlockchain", t, func() {

		Convey(".Init", func() {
			err := pgChain.Init()
			So(err, ShouldBeNil)
			chainTableExists := db.(*gorm.DB).HasTable(ChainTableName)
			So(chainTableExists, ShouldEqual, true)
			blockTableExists := db.(*gorm.DB).HasTable(BlockTableName)
			So(blockTableExists, ShouldEqual, true)
		})

		Convey(".MakeChainName", func() {
			expected := fmt.Sprintf("%s;%s", "cocooncode_1", "accounts")
			So(expected, ShouldEqual, pgChain.MakeChainName("cocooncode_1", "accounts"))
		})

		Convey(".CreateChain", func() {

			Convey("Should successfully create a chain", func() {
				chainName := util.RandString(5)
				chain, err := pgChain.CreateChain(chainName, true)
				So(err, ShouldBeNil)
				So(chain.Name, ShouldEqual, chainName)
				So(chain.Public, ShouldEqual, true)

				Convey("Should fail when trying to create a chain with an already used name", func() {
					chain, err := pgChain.CreateChain(chainName, true)
					So(chain, ShouldBeNil)
					So(err, ShouldNotBeNil)
					So(err.Error(), ShouldEqual, "chain with matching name already exists")
				})

				Convey(".GetChain", func() {

					Convey("Should successfully return existing chain", func() {
						chain, err := pgChain.GetChain(chainName)
						So(err, ShouldBeNil)
						So(chain.Name, ShouldEqual, chainName)
						So(chain.Public, ShouldEqual, true)
					})

					Convey("Should return nil, nil result if chain does not exists", func() {
						chain, err := pgChain.GetChain("chain2")
						So(chain, ShouldBeNil)
						So(err, ShouldBeNil)
					})
				})
			})
		})

		Convey(".MakeTxsHash", func() {

			Convey("Should successfully return expected sha256 hash", func() {
				txs := []*types.Transaction{
					&types.Transaction{Hash: util.Sha256("a")},
					&types.Transaction{Hash: util.Sha256("b")},
					&types.Transaction{Hash: util.Sha256("c")},
				}
				hash := MakeTxsHash(txs)
				So(len(hash), ShouldEqual, 64)
				So(hash, ShouldEqual, "a6309e827156ce1b9b532024b811f2966274d12794640a7ac7c66e84c5e2e9ae")
			})
		})

		Convey(".VerifyTxs", func() {

			Convey("Should successfully verify all transactions to be accurate", func() {
				tx1 := &types.Transaction{Number: 1, Ledger: "ledger1", ID: "some_id", Key: "key", Value: "value", CreatedAt: 123456789}
				tx1.Hash = tx1.MakeHash()
				tx2 := &types.Transaction{Number: 2, Ledger: "ledger2", ID: "some_id", Key: "key", Value: "value", CreatedAt: 123499789}
				tx2.Hash = tx2.MakeHash()
				txs := []*types.Transaction{
					tx1,
					tx2,
				}
				failedTx, verified := VerifyTxs(txs)
				So(failedTx, ShouldBeNil)
				So(verified, ShouldEqual, true)
			})

			Convey("Should fail if at least one tx hash is invalid", func() {
				tx1 := &types.Transaction{Number: 1, Ledger: "ledger1", ID: "some_id", Key: "key", Value: "value", CreatedAt: 123456789}
				tx1.Hash = tx1.MakeHash()
				tx2 := &types.Transaction{Number: 2, Ledger: "ledger2", ID: "some_id", Key: "key", Value: "value", CreatedAt: 123499789}
				tx2.Hash = tx2.MakeHash()
				tx3 := &types.Transaction{Number: 3, Ledger: "ledger3", ID: "some_id", Key: "key", Value: "value", CreatedAt: 123499789}
				tx3.Hash = "very very wrong hash"
				txs := []*types.Transaction{
					tx1,
					tx2,
					tx3,
				}
				failedTx, verified := VerifyTxs(txs)
				So(failedTx, ShouldNotBeNil)
				So(verified, ShouldEqual, false)
				So(tx3, ShouldResemble, failedTx)
			})
		})

		Convey(".CreateBlock", func() {

			chainName := util.RandString(5)
			chain, err := pgChain.CreateChain(chainName, true)
			So(err, ShouldBeNil)
			So(chain.Name, ShouldEqual, chainName)
			So(chain.Public, ShouldEqual, true)

			Convey("Should return error if chain does not exist", func() {
				txs := []*types.Transaction{{ID: util.UUID4()}}
				blk, err := pgChain.CreateBlock(util.RandString(10), "unknown", txs)
				So(blk, ShouldBeNil)
				So(err, ShouldEqual, types.ErrChainNotFound)
			})

			Convey("Should return error if no transaction is provided", func() {
				blk, err := pgChain.CreateBlock(util.RandString(10), chainName, []*types.Transaction{})
				So(blk, ShouldBeNil)
				So(err, ShouldEqual, types.ErrZeroTransactions)
			})

			Convey("Should return error if a transaction hash is invalid", func() {
				tx1 := &types.Transaction{Number: 1, Ledger: "ledger1", ID: "some_id", Key: "key", Value: "value", CreatedAt: 123456789}
				tx1.Hash = tx1.MakeHash()
				tx2 := &types.Transaction{Number: 2, Ledger: "ledger2", ID: "some_id", Key: "key", Value: "value", CreatedAt: 123499789}
				tx2.Hash = "wrong hash"
				txs := []*types.Transaction{tx1, tx2}
				blk, err := pgChain.CreateBlock(util.RandString(10), chainName, txs)
				So(blk, ShouldBeNil)
				So(err.Error(), ShouldContainSubstring, "has an invalid hash")
			})

			Convey("Should successfully create the first block with expected block values", func() {
				tx1 := &types.Transaction{Number: 1, Ledger: "ledger1", ID: "some_id", Key: "key", Value: "value", CreatedAt: 123456789}
				tx1.Hash = tx1.MakeHash()
				tx2 := &types.Transaction{Number: 2, Ledger: "ledger2", ID: "some_id", Key: "key", Value: "value", CreatedAt: 123499789}
				tx2.Hash = tx2.MakeHash()
				txs := []*types.Transaction{tx1, tx2}
				blk, err := pgChain.CreateBlock(util.RandString(10), chainName, txs)
				So(blk, ShouldNotBeNil)
				So(err, ShouldBeNil)
				So(blk.ChainName, ShouldEqual, chainName)
				So(blk.Number, ShouldEqual, 1)
				So(blk.PrevBlockHash, ShouldEqual, MakeGenesisBlockHash(chainName))
				So(blk.Hash, ShouldEqual, MakeTxsHash(txs))
				txsBytes, _ := util.ToJSON(txs)
				So(blk.Transactions, ShouldResemble, txsBytes)

				Convey("Should successfully add another block that references the previous block", func() {
					tx1 := &types.Transaction{Number: 1, Ledger: "ledger1", ID: "some_id", Key: "key", Value: "value", CreatedAt: 123456789}
					tx1.Hash = tx1.MakeHash()
					txs := []*types.Transaction{tx1}

					blk2, err := pgChain.CreateBlock(util.RandString(10), chainName, txs)
					So(blk2, ShouldNotBeNil)
					So(err, ShouldBeNil)
					So(blk2.ChainName, ShouldEqual, chainName)
					So(blk2.Number, ShouldEqual, 2)
					So(blk2.PrevBlockHash, ShouldEqual, blk.Hash)
					So(blk2.Hash, ShouldEqual, MakeTxsHash(txs))
					txsBytes, _ := util.ToJSON(txs)
					So(blk2.Transactions, ShouldResemble, txsBytes)
				})
			})
		})

		Convey(".GetBlock", func() {

			chainName := util.RandString(5)
			chain, err := pgChain.CreateChain(chainName, true)
			So(err, ShouldBeNil)
			So(chain.Name, ShouldEqual, chainName)
			So(chain.Public, ShouldEqual, true)

			Convey("Should return error nil and nil if block does not exists", func() {
				block, err := pgChain.GetBlock(chain.Name, "unknown_id")
				So(block, ShouldBeNil)
				So(err, ShouldBeNil)
			})

			Convey("Should successfully return an existing block", func() {

				tx1 := &types.Transaction{Number: 1, Ledger: "ledger1", ID: "some_id_2", Key: "key", Value: "value", CreatedAt: 123456789}
				tx1.Hash = tx1.MakeHash()
				txs := []*types.Transaction{tx1}
				blk, err := pgChain.CreateBlock(util.RandString(10), chain.Name, txs)
				So(err, ShouldBeNil)
				So(blk, ShouldNotBeNil)

				block, err := pgChain.GetBlock(chain.Name, blk.ID)
				So(err, ShouldBeNil)
				So(block, ShouldNotBeNil)
				So(block.ID, ShouldEqual, blk.ID)
				So(block.Hash, ShouldEqual, blk.Hash)
			})
		})
	})
}
