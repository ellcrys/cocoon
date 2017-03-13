package orderer

import (
	"testing"
	"time"

	"os"

	"github.com/ellcrys/util"
	"github.com/ncodes/cocoon/core/orderer/proto"
	"github.com/ncodes/cocoon/core/store/impl"
	"github.com/ncodes/cocoon/core/types"
	. "github.com/smartystreets/goconvey/convey" // convey needs this
	context "golang.org/x/net/context"
)

var storeConStr = util.Env("STORE_CON_STR", "host=localhost user=ned dbname=cocoon_dev sslmode=disable password=")

func init() {
	os.Setenv("APP_ENV", "test")
}

func startOrderer(startCB func(*Orderer, chan bool)) {
	endCh := make(chan bool)
	addr := util.Env("ORDERER_ADDR", "127.0.0.1:7001")
	newOrderer := NewOrderer()
	newOrderer.SetStore(new(impl.PostgresStore))
	go newOrderer.Start(addr, storeConStr, endCh)
	time.Sleep(3 * time.Second)
	startCB(newOrderer, endCh)
	<-endCh
}

func dropStoreDB() error {
	return impl.Destroy(storeConStr)
}

func TestOrderer(t *testing.T) {

	err := dropStoreDB()
	if err != nil {
		t.Log("failed to drop chain")
		t.Fail()
	}

	startOrderer(func(od *Orderer, endCh chan bool) {

		Convey("Orderer", t, func() {

			Convey(".CreateLedger", func() {

				Convey("Should create a ledger", func() {
					ledger, err := od.CreateLedger(context.Background(), &proto.CreateLedgerParams{
						CocoonCodeId: "cocoon-123",
						Name:         "myledger",
						Public:       true,
					})
					So(err, ShouldBeNil)
					So(ledger, ShouldNotBeNil)
					So(ledger.Name, ShouldEqual, "d334aae9e18f2643bb7da8ed723e07b79b25f673e3f5e5be5f258512de5fc9e9")

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
						So(ledger.Name, ShouldEqual, "d334aae9e18f2643bb7da8ed723e07b79b25f673e3f5e5be5f258512de5fc9e9")
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
							Id:           util.UUID4(),
							CocoonCodeId: "cocoon-123",
							LedgerName:   "unknown",
							Key:          "mykey",
							Value:        []byte("value"),
						})
						So(tx, ShouldBeNil)
						So(err, ShouldNotBeNil)
						So(err, ShouldResemble, types.ErrLedgerNotFound)
					})

					Convey("Should successfully put a transaction in a ledger", func() {

						ledgerName := util.UUID4()
						ledger, err := od.CreateLedger(context.Background(), &proto.CreateLedgerParams{
							CocoonCodeId: "cocoon-abc",
							Name:         ledgerName,
							Public:       true,
						})
						So(err, ShouldBeNil)
						So(ledger, ShouldNotBeNil)

						tx, err := od.Put(context.Background(), &proto.PutTransactionParams{
							Id:           util.UUID4(),
							CocoonCodeId: "cocoon-abc",
							LedgerName:   ledgerName,
							Key:          "mykey",
							Value:        []byte("value"),
						})

						So(tx, ShouldNotBeNil)
						So(err, ShouldBeNil)

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
						tx, err := od.Put(context.Background(), &proto.PutTransactionParams{
							Id:           util.UUID4(),
							CocoonCodeId: "cocoon-abc",
							LedgerName:   ledgerName,
							Key:          key,
							Value:        []byte("value"),
						})

						t.Log(tx == nil, err)
						So(err, ShouldBeNil)

						Convey("Should return error if ledger does not exist", func() {
							t.Log(time.Now().Unix())
							tx, err := od.Get(context.Background(), &proto.GetParams{
								CocoonCodeId: "cocoon-abc",
								Ledger:       "unknown",
								Key:          "wrong",
							})
							So(tx, ShouldBeNil)
							So(err, ShouldResemble, types.ErrLedgerNotFound)
						})

						Convey("Should return transaction error if key does not exist in ledger", func() {
							t.Log(time.Now().Unix())
							tx, err := od.Get(context.Background(), &proto.GetParams{
								CocoonCodeId: "cocoon-abc",
								Ledger:       ledgerName,
								Key:          "wrong",
							})
							So(tx, ShouldBeNil)
							So(err, ShouldResemble, types.ErrTxNotFound)
						})

						Convey("Should successfully get a transaction", func() {
							t.Log(time.Now().Unix())
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
	})
}
