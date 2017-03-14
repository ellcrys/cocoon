package orderer

import (
	"context"
	"strings"
	"testing"
	"time"

	"os"

	"os/exec"

	"fmt"

	"github.com/ellcrys/util"
	blkch_impl "github.com/ncodes/cocoon/core/blockchain/impl"
	"github.com/ncodes/cocoon/core/orderer/proto"
	"github.com/ncodes/cocoon/core/store/impl"
	"github.com/ncodes/cocoon/core/types"
	. "github.com/smartystreets/goconvey/convey" // convey needs this
)

var dbname = "test_db_" + util.RandString(5)
var storeConStr = util.Env("STORE_CON_STR", "host=localhost user=ned dbname="+dbname+" sslmode=disable password=")

func init() {
	os.Setenv("APP_ENV", "test")
}

func createDb(t *testing.T) error {
	if err := exec.Command("createdb", dbname).Start(); err != nil {
		return fmt.Errorf("failed to create test db")
	}
	return nil
}

func dropDB(t *testing.T) error {
	if err := exec.Command("dropdb", dbname).Start(); err != nil {
		return fmt.Errorf("failed to drop test db")
	}
	return impl.Destroy(storeConStr)
}

func startOrderer(startCB func(*Orderer, chan bool)) {
	endCh := make(chan bool)
	addr := util.Env("ORDERER_ADDR", "127.0.0.1:7001")
	newOrderer := NewOrderer()
	newOrderer.SetStore(new(impl.PostgresStore))
	newOrderer.SetBlockchain(new((blkch_impl.PostgresBlockchain)))
	go newOrderer.Start(addr, storeConStr, endCh)
	time.Sleep(3 * time.Second)
	startCB(newOrderer, endCh)
	<-endCh
}

func TestOrderer(t *testing.T) {

	err := createDb(t)
	if err != nil {
		t.Fatal(err)
	}

	startOrderer(func(od *Orderer, endCh chan bool) {

		Convey("Orderer", t, func() {

			Convey(".CreateLedger", func() {

				Convey("Should create a ledger and chain", func() {
					ledger, err := od.CreateLedger(context.Background(), &proto.CreateLedgerParams{
						CocoonCodeId: "cocoon-123",
						Name:         "myledger",
						Public:       true,
						Chained:      true,
					})
					So(err, ShouldBeNil)
					So(ledger, ShouldNotBeNil)
					So(ledger.Name, ShouldEqual, "myledger")
					So(ledger.Chained, ShouldEqual, true)

					chain, err := od.blockchain.GetChain(od.store.MakeLedgerName("cocoon-123", "myledger"))
					So(err, ShouldBeNil)
					So(chain, ShouldNotBeNil)

					Convey("Should return error if ledger with same name exists", func() {
						ledger, err := od.CreateLedger(context.Background(), &proto.CreateLedgerParams{
							CocoonCodeId: "cocoon-123",
							Name:         "myledger",
							Public:       false,
						})
						So(err, ShouldNotBeNil)
						So(err.Error(), ShouldEqual, "ledger with matching name already exists")
						So(ledger, ShouldBeNil)
					})
				})

				Convey(".GetLedger", func() {
					Convey("Should successfully return ledger with matching name", func() {
						ledger, err := od.GetLedger(context.Background(), &proto.GetLedgerParams{
							CocoonCodeId: "cocoon-123",
							Name:         "myledger",
						})
						So(err, ShouldBeNil)
						So(ledger, ShouldNotBeNil)
						So(ledger.Name, ShouldEqual, "myledger")
						So(ledger.Chained, ShouldEqual, true)
					})

					Convey("Should return error if no legder with matching cocoon id and name", func() {
						ledger, err := od.GetLedger(context.Background(), &proto.GetLedgerParams{
							CocoonCodeId: "cocoon-124",
							Name:         "myledger",
						})
						So(err, ShouldResemble, types.ErrLedgerNotFound)
						So(ledger, ShouldBeNil)
					})
				})

				Convey(".Put", func() {

					Convey("Should return error if ledger does not exists", func() {

						tx, err := od.Put(context.Background(), &proto.PutTransactionParams{
							CocoonCodeId: "cocoon-123",
							LedgerName:   "unknown",
							Transactions: []*proto.Transaction{},
						})
						So(tx, ShouldBeNil)
						So(err, ShouldNotBeNil)
						So(err, ShouldResemble, types.ErrLedgerNotFound)
					})

					Convey("Should successfully put transactions in a ledger", func() {

						ledgerName := util.UUID4()
						ledger, err := od.CreateLedger(context.Background(), &proto.CreateLedgerParams{
							CocoonCodeId: "cocoon-abc",
							Name:         ledgerName,
							Public:       true,
						})
						So(err, ShouldBeNil)
						So(ledger, ShouldNotBeNil)

						txs := []*proto.Transaction{
							{Ledger: ledgerName, Id: util.Sha256(util.UUID4()), Key: util.Sha256(util.UUID4()), Value: util.Sha256(util.UUID4())},
							{Ledger: ledgerName, Id: util.Sha256(util.UUID4()), Key: util.Sha256(util.UUID4()), Value: util.Sha256(util.UUID4())},
						}

						tx, err := od.Put(context.Background(), &proto.PutTransactionParams{
							CocoonCodeId: "cocoon-abc",
							LedgerName:   ledgerName,
							Transactions: txs,
						})

						So(err, ShouldBeNil)
						So(tx, ShouldNotBeNil)
						So(tx.Added, ShouldEqual, 2)
						So(*tx.Block, ShouldResemble, proto.Block{})
					})

					Convey("Should successfully put transactions in a chained ledger and create a block", func() {
						ledgerName := util.UUID4()
						ledger, err := od.CreateLedger(context.Background(), &proto.CreateLedgerParams{
							CocoonCodeId: "cocoon-abc",
							Name:         ledgerName,
							Public:       true,
							Chained:      true,
						})
						So(err, ShouldBeNil)
						So(ledger, ShouldNotBeNil)
						So(ledger.Chained, ShouldEqual, true)

						txs := []*proto.Transaction{
							{Ledger: ledgerName, Id: util.Sha256(util.UUID4()), Key: util.Sha256(util.UUID4()), Value: util.Sha256(util.UUID4())},
							{Ledger: ledgerName, Id: util.Sha256(util.UUID4()), Key: util.Sha256(util.UUID4()), Value: util.Sha256(util.UUID4())},
						}

						tx, err := od.Put(context.Background(), &proto.PutTransactionParams{
							CocoonCodeId: "cocoon-abc",
							LedgerName:   ledgerName,
							Transactions: txs,
						})

						So(err, ShouldBeNil)
						So(tx, ShouldNotBeNil)
						So(tx.Added, ShouldEqual, 2)
						So(tx.Block, ShouldNotBeNil)
						So(tx.Block.PrevBlockHash, ShouldEqual, strings.Repeat("0", 64))
						So(tx.Block.Number, ShouldEqual, 1)
					})

					Convey(".Get", func() {

						ledgerName := util.UUID4()
						ledger, err := od.CreateLedger(context.Background(), &proto.CreateLedgerParams{
							CocoonCodeId: "cocoon-abc",
							Name:         ledgerName,
							Public:       true,
						})
						So(err, ShouldBeNil)
						So(ledger, ShouldNotBeNil)

						key := util.UUID4()
						txs := []*proto.Transaction{
							{Ledger: ledgerName, Id: util.Sha256(util.UUID4()), Key: key, Value: util.Sha256(util.UUID4())},
						}

						tx, err := od.Put(context.Background(), &proto.PutTransactionParams{
							CocoonCodeId: "cocoon-abc",
							LedgerName:   ledgerName,
							Transactions: txs,
						})
						So(tx, ShouldNotBeNil)
						So(err, ShouldBeNil)

						Convey("Should return error if ledger does not exist", func() {
							tx, err := od.Get(context.Background(), &proto.GetParams{
								CocoonCodeId: "cocoon-abc",
								Ledger:       "unknown",
								Key:          "wrong",
							})
							So(tx, ShouldBeNil)
							So(err, ShouldResemble, types.ErrLedgerNotFound)
						})

						Convey("Should return transaction error if key does not exist in ledger", func() {
							tx, err := od.Get(context.Background(), &proto.GetParams{
								CocoonCodeId: "cocoon-abc",
								Ledger:       ledgerName,
								Key:          "wrong",
							})
							So(tx, ShouldBeNil)
							So(err, ShouldResemble, types.ErrTxNotFound)
						})

						Convey("Should successfully get a transaction", func() {
							tx, err := od.Get(context.Background(), &proto.GetParams{
								CocoonCodeId: "cocoon-abc",
								Ledger:       ledgerName,
								Key:          key,
							})
							So(tx, ShouldNotBeNil)
							So(err, ShouldBeNil)
						})
					})
				})
			})
		})

		close(endCh)
		err := dropDB(t)
		if err != nil {
			t.Fail()
		}
	})
}
