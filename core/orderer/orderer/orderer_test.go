package orderer

import (
	"testing"
	"time"

	"os"

	"os/exec"

	"fmt"

	"github.com/ellcrys/util"
	blkch_impl "github.com/ncodes/cocoon/core/blockchain/impl"
	"github.com/ncodes/cocoon/core/orderer/proto_orderer"
	"github.com/ncodes/cocoon/core/store/impl"
	"github.com/ncodes/cocoon/core/types"
	logging "github.com/op/go-logging"
	. "github.com/smartystreets/goconvey/convey" // convey needs this
	context "golang.org/x/net/context"
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
	SetLogLevel(logging.CRITICAL)
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
					ledger, err := od.CreateLedger(context.Background(), &proto_orderer.CreateLedgerParams{
						CocoonID: "cocoon-123",
						Name:     "myledger",
						Public:   true,
						Chained:  true,
					})
					So(err, ShouldBeNil)
					So(ledger, ShouldNotBeNil)
					So(ledger.Name, ShouldEqual, "myledger")
					So(ledger.Chained, ShouldEqual, true)

					chain, err := od.blockchain.GetChain(types.MakeLedgerName("cocoon-123", "myledger"))
					So(err, ShouldBeNil)
					So(chain, ShouldNotBeNil)

					Convey("Should return error if ledger with same name exists", func() {
						ledger, err := od.CreateLedger(context.Background(), &proto_orderer.CreateLedgerParams{
							CocoonID: "cocoon-123",
							Name:     "myledger",
							Public:   false,
						})
						So(err, ShouldNotBeNil)
						So(err.Error(), ShouldEqual, "ledger with matching name already exists")
						So(ledger, ShouldBeNil)
					})
				})

				Convey(".GetLedger", func() {

					ledgerName := util.RandString(10)
					ledger, err := od.CreateLedger(context.Background(), &proto_orderer.CreateLedgerParams{
						CocoonID: "cocoon-123",
						Name:     ledgerName,
						Public:   true,
						Chained:  true,
					})
					So(err, ShouldBeNil)
					So(ledger, ShouldNotBeNil)

					Convey("Should successfully return ledger with matching name", func() {
						ledger, err := od.GetLedger(context.Background(), &proto_orderer.GetLedgerParams{
							CocoonID: "cocoon-123",
							Name:     ledgerName,
						})
						So(err, ShouldBeNil)
						So(ledger, ShouldNotBeNil)
						So(ledger.Name, ShouldEqual, ledgerName)
						So(ledger.Chained, ShouldEqual, true)
					})

					Convey("Should return error if no legder with matching cocoon id and name", func() {
						ledger, err := od.GetLedger(context.Background(), &proto_orderer.GetLedgerParams{
							CocoonID: "cocoon-124",
							Name:     "myledger",
						})
						So(err, ShouldResemble, types.ErrLedgerNotFound)
						So(ledger, ShouldBeNil)
					})

					Reset(func() {
						impl.Clear(storeConStr)
					})
				})

				Convey(".Put", func() {

					Convey("Should return error if ledger does not exists", func() {

						tx, err := od.Put(context.Background(), &proto_orderer.PutTransactionParams{
							CocoonID:     "cocoon-123",
							LedgerName:   "unknown",
							Transactions: []*proto_orderer.Transaction{},
						})
						So(tx, ShouldBeNil)
						So(err, ShouldNotBeNil)
						So(err, ShouldResemble, types.ErrLedgerNotFound)
					})

					Convey("Should successfully put transactions in a ledger", func() {

						ledgerName := util.UUID4()
						ledger, err := od.CreateLedger(context.Background(), &proto_orderer.CreateLedgerParams{
							CocoonID: "cocoon-abc",
							Name:     ledgerName,
							Public:   true,
						})
						So(err, ShouldBeNil)
						So(ledger, ShouldNotBeNil)

						txs := []*proto_orderer.Transaction{
							{Ledger: ledgerName, Id: util.Sha256(util.UUID4()), Key: util.Sha256(util.UUID4()), Value: util.Sha256(util.UUID4())},
							{Ledger: ledgerName, Id: util.Sha256(util.UUID4()), Key: util.Sha256(util.UUID4()), Value: util.Sha256(util.UUID4())},
						}

						result, err := od.Put(context.Background(), &proto_orderer.PutTransactionParams{
							CocoonID:     "cocoon-abc",
							LedgerName:   ledgerName,
							Transactions: txs,
						})

						So(err, ShouldBeNil)
						So(result, ShouldNotBeNil)
						So(len(result.TxReceipts), ShouldEqual, 2)
						So(result.TxReceipts[0].ID, ShouldEqual, txs[0].Id)
						So(result.TxReceipts[1].ID, ShouldEqual, txs[1].Id)
						So(result.Block, ShouldBeNil)
					})

					Convey("Should successfully put transactions in a chained ledger and create a block", func() {
						chainedLedgerName := util.UUID4()
						chainedLedger, err := od.CreateLedger(context.Background(), &proto_orderer.CreateLedgerParams{
							CocoonID: "cocoon-abc",
							Name:     chainedLedgerName,
							Public:   true,
							Chained:  true,
						})
						So(err, ShouldBeNil)
						So(chainedLedger, ShouldNotBeNil)
						So(chainedLedger.Chained, ShouldEqual, true)

						txs := []*proto_orderer.Transaction{
							{Ledger: chainedLedgerName, Id: util.Sha256(util.UUID4()), Key: util.Sha256(util.UUID4()), Value: util.Sha256(util.UUID4())},
							{Ledger: chainedLedgerName, Id: util.Sha256(util.UUID4()), Key: util.Sha256(util.UUID4()), Value: util.Sha256(util.UUID4())},
						}

						result, err := od.Put(context.Background(), &proto_orderer.PutTransactionParams{
							CocoonID:     "cocoon-abc",
							LedgerName:   chainedLedger.Name,
							Transactions: txs,
						})

						So(err, ShouldBeNil)
						So(result, ShouldNotBeNil)
						So(len(result.TxReceipts), ShouldEqual, 2)
						So(result.TxReceipts[0].ID, ShouldEqual, txs[0].Id)
						So(result.TxReceipts[1].ID, ShouldEqual, txs[1].Id)
						So(result.Block, ShouldNotBeNil)
						So(result.Block.PrevBlockHash, ShouldEqual, blkch_impl.MakeGenesisBlockHash(types.MakeLedgerName("cocoon-abc", chainedLedgerName)))
						So(result.Block.Number, ShouldEqual, 1)
					})
				})

				Convey(".Get", func() {

					ledgerName := util.UUID4()
					ledger, err := od.CreateLedger(context.Background(), &proto_orderer.CreateLedgerParams{
						CocoonID: "cocoon-abc",
						Name:     ledgerName,
						Public:   true,
					})
					So(err, ShouldBeNil)
					So(ledger, ShouldNotBeNil)

					key := util.UUID4()
					id := util.Sha256(util.UUID4())
					txs := []*proto_orderer.Transaction{
						{Ledger: ledgerName, Id: id, Key: key, Value: util.Sha256(util.UUID4())},
					}

					tx, err := od.Put(context.Background(), &proto_orderer.PutTransactionParams{
						CocoonID:     "cocoon-abc",
						LedgerName:   ledgerName,
						Transactions: txs,
					})
					So(tx, ShouldNotBeNil)
					So(err, ShouldBeNil)

					Convey("Should return error if ledger does not exist", func() {
						tx, err := od.Get(context.Background(), &proto_orderer.GetParams{
							CocoonID: "cocoon-abc",
							Ledger:   "unknown",
							Key:      "wrong",
						})
						So(tx, ShouldBeNil)
						So(err, ShouldResemble, types.ErrLedgerNotFound)
					})

					Convey("Should return transaction error if key does not exist in ledger", func() {
						tx, err := od.Get(context.Background(), &proto_orderer.GetParams{
							CocoonID: "cocoon-abc",
							Ledger:   ledgerName,
							Key:      "wrong",
						})
						So(tx, ShouldBeNil)
						So(err, ShouldResemble, types.ErrTxNotFound)
					})

					Convey("Should successfully get a transaction", func() {
						tx, err := od.Get(context.Background(), &proto_orderer.GetParams{
							CocoonID: "cocoon-abc",
							Ledger:   ledgerName,
							Key:      key,
						})
						So(tx, ShouldNotBeNil)
						So(err, ShouldBeNil)

						Convey(".GetByID", func() {

							Convey("Should return error if transaction is not found", func() {
								tx, err := od.GetByID(context.Background(), &proto_orderer.GetParams{
									CocoonID: "cocoon-abc",
									Ledger:   ledgerName,
									Id:       "unknown-123",
								})
								So(tx, ShouldBeNil)
								So(err, ShouldEqual, types.ErrTxNotFound)
							})

							Convey("Should successfully get a transaction by it's id", func() {
								tx, err := od.GetByID(context.Background(), &proto_orderer.GetParams{
									CocoonID: "cocoon-abc",
									Ledger:   ledgerName,
									Id:       id,
								})
								So(tx, ShouldNotBeNil)
								So(err, ShouldBeNil)
								So(tx.Id, ShouldEqual, id)
							})
						})
					})
				})

				Convey(".GetBlockByID", func() {
					ledgerName := util.UUID4() + "-okay"
					ledger, err := od.CreateLedger(context.Background(), &proto_orderer.CreateLedgerParams{
						CocoonID: "cocoon-abc",
						Name:     ledgerName,
						Public:   true,
						Chained:  true,
					})
					So(err, ShouldBeNil)
					So(ledger, ShouldNotBeNil)

					key := util.UUID4()
					id := util.Sha256(util.UUID4())
					txs := []*proto_orderer.Transaction{
						{Ledger: ledgerName, Id: id, Key: key, Value: util.Sha256(util.UUID4())},
					}

					result, err := od.Put(context.Background(), &proto_orderer.PutTransactionParams{
						CocoonID:     "cocoon-abc",
						LedgerName:   ledgerName,
						Transactions: txs,
					})

					So(result, ShouldNotBeNil)
					So(err, ShouldBeNil)
					So(result.Block, ShouldNotBeNil)
					So(len(result.TxReceipts), ShouldEqual, 1)
					So(result.TxReceipts[0].ID, ShouldEqual, txs[0].Id)

					block, err := od.GetBlockByID(context.Background(), &proto_orderer.GetBlockParams{
						CocoonID: "cocoon-abc",
						Ledger:   ledgerName,
						Id:       result.Block.Id,
					})

					So(err, ShouldBeNil)
					So(block, ShouldNotBeNil)
					So(block.Id, ShouldEqual, result.Block.Id)
					So(block.Hash, ShouldEqual, result.Block.Hash)
				})

				Convey(".GetRange", func() {

					ledgerName := util.UUID4()
					ledger, err := od.CreateLedger(context.Background(), &proto_orderer.CreateLedgerParams{
						CocoonID: "cocoon-abc",
						Name:     ledgerName,
						Public:   true,
						Chained:  true,
					})
					So(err, ShouldBeNil)
					So(ledger, ShouldNotBeNil)

					txs := []*proto_orderer.Transaction{
						{Id: util.RandString(10), Key: "account.ken", Value: util.Sha256(util.UUID4())},
						{Id: util.RandString(10), Key: "account.glen", Value: util.Sha256(util.UUID4())},
						{Id: util.RandString(10), Key: "x", Value: util.Sha256(util.UUID4())},
					}

					result, err := od.Put(context.Background(), &proto_orderer.PutTransactionParams{
						CocoonID:     "cocoon-abc",
						LedgerName:   ledgerName,
						Transactions: txs,
					})

					So(result, ShouldNotBeNil)
					So(err, ShouldBeNil)

					Convey("Should return exepected transaction range with inclusive option disabled", func() {
						txs, err := od.GetRange(context.Background(), &proto_orderer.GetRangeParams{
							CocoonID:  "cocoon-abc",
							Ledger:    ledgerName,
							StartKey:  "account",
							EndKey:    "x",
							Inclusive: false,
							Limit:     10,
						})
						So(txs, ShouldNotBeNil)
						So(err, ShouldBeNil)
						So(len(txs.Transactions), ShouldEqual, 2)
					})

					Convey("Should return exepected transaction range with inclusive option enabled", func() {
						txs, err := od.GetRange(context.Background(), &proto_orderer.GetRangeParams{
							CocoonID:  "cocoon-abc",
							Ledger:    ledgerName,
							StartKey:  "account",
							EndKey:    "x",
							Inclusive: true,
							Limit:     10,
						})

						So(txs, ShouldNotBeNil)
						So(err, ShouldBeNil)
						So(len(txs.Transactions), ShouldEqual, 3)
					})

					Reset(func() {
						impl.Clear(storeConStr)
					})
				})

				Reset(func() {
					impl.Clear(storeConStr)
				})
			})

			Reset(func() {
				impl.Clear(storeConStr)
			})
		})

		close(endCh)
		err := dropDB(t)
		if err != nil {
			t.Fail()
		}
	})
}
