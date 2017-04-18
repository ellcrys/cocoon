package api

import (
	"testing"
	"time"

	"google.golang.org/grpc/metadata"

	"os"

	"os/exec"

	"fmt"

	"github.com/ellcrys/util"
	"github.com/ncodes/cocoon/core/api/api/proto"
	blkch_impl "github.com/ncodes/cocoon/core/blockchain/impl"
	"github.com/ncodes/cocoon/core/orderer/orderer"
	"github.com/ncodes/cocoon/core/scheduler"
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

func startOrderer(startCB func(*orderer.Orderer, chan bool)) {
	endCh := make(chan bool)
	os.Setenv("DEV_ORDERER_ADDR", "127.0.0.1:7011")
	addr := util.Env("DEV_ORDERER_ADDR", "")
	orderer.SetLogLevel(logging.CRITICAL)
	newOrderer := orderer.NewOrderer()
	newOrderer.SetStore(new(impl.PostgresStore))
	newOrderer.SetBlockchain(new((blkch_impl.PostgresBlockchain)))
	go newOrderer.Start(addr, storeConStr, endCh)
	startCB(newOrderer, endCh)
	<-endCh
}

func startAPIServer(startCB func(*API, chan bool)) {
	endCh := make(chan bool)
	apiServer := NewAPI(scheduler.NewNomad())
	addr := util.Env("API_ADDRESS", "127.0.0.1:7004")
	go apiServer.Start(addr, endCh)
	time.Sleep(3 * time.Second)
	startCB(apiServer, endCh)
	<-endCh
}

func TestOrderer(t *testing.T) {

	err := createDb(t)
	if err != nil {
		t.Fatal(err)
	}

	key := "secret"

	startOrderer(func(od *orderer.Orderer, endCh chan bool) {
		startAPIServer(func(api *API, apiEndCh chan bool) {
			Convey("API", t, func() {

				Convey(".CreateIdentity", func() {

					Convey("Should successfully create an identity", func() {
						identity := types.Identity{
							Email:    util.RandString(5) + "@example.com",
							Password: "somepassword",
						}

						i, err := api.CreateIdentity(context.Background(), &proto.CreateIdentityRequest{
							Email:    identity.Email,
							Password: identity.Password,
						})
						So(err, ShouldBeNil)
						So(i.Status, ShouldEqual, 200)

						Convey("Should return error if identity with same credentials already exists", func() {
							i, err := api.CreateIdentity(context.Background(), &proto.CreateIdentityRequest{
								Email:    identity.Email,
								Password: identity.Password,
							})
							So(i, ShouldBeNil)
							So(err, ShouldNotBeNil)
							So(err.Error(), ShouldEqual, "an identity with matching email already exists")
						})

						Convey(".CreateCocoon", func() {

							ss, err := makeAuthToken(util.UUID4(), identity.GetID(), "token.cli", time.Now().AddDate(0, 1, 0).Unix(), key)
							So(err, ShouldBeNil)
							md := metadata.Pairs("access_token", ss)
							ctx := context.Background()
							ctx = metadata.NewContext(ctx, md)

							Convey("Should successfully create a cocoon", func() {

								id := util.UUID4()
								r, err := api.CreateCocoon(ctx, &proto.CocoonPayloadRequest{
									ID:             id,
									URL:            "https://github.com/ncodes/cocoon-example-01",
									Language:       "go",
									Memory:         "512m",
									NumSignatories: 1,
									SigThreshold:   1,
									CPUShares:      "1x",
								})
								So(err, ShouldBeNil)
								So(r.Status, ShouldEqual, 200)
								So(len(r.Body), ShouldNotEqual, 0)

								Convey("Should fail to create cocoon with an already used id", func() {
									r, err := api.CreateCocoon(ctx, &proto.CocoonPayloadRequest{
										ID:             id,
										URL:            "https://github.com/ncodes/cocoon-example-01",
										Language:       "go",
										Memory:         "512m",
										CPUShares:      "1x",
										NumSignatories: 1,
										SigThreshold:   1,
									})
									So(r, ShouldBeNil)
									So(err, ShouldNotBeNil)
									So(err.Error(), ShouldEqual, "cocoon with matching ID already exists")
								})

								Convey(".GetCocoon", func() {
									Convey("Should successfully get an existing cocoon by id", func() {
										c, err := api.GetCocoon(context.Background(), &proto.GetCocoonRequest{
											ID: id,
										})
										So(err, ShouldBeNil)
										So(c.Status, ShouldEqual, 200)
									})

									Convey("Should return error if cocoon does not exists", func() {
										_, err = api.GetCocoon(context.Background(), &proto.GetCocoonRequest{
											ID: "unknown",
										})
										So(err, ShouldNotBeNil)
										So(err.Error(), ShouldEqual, "cocoon not found")
									})
								})
							})
						})

						Convey(".CreateRelease", func() {

							Convey("Should successfully create a release", func() {
								id := util.UUID4()
								r, err := api.CreateRelease(context.Background(), &proto.CreateReleaseRequest{
									ID:       id,
									CocoonID: "cocoon-123",
									URL:      "https://github.com/ncodes/cocoon-example-01",
									Language: supportedLanguages[0],
								})
								So(err, ShouldBeNil)
								So(r.Status, ShouldEqual, 200)
								So(len(r.Body), ShouldNotEqual, 0)

								Convey("Should fail to create release with an already used id", func() {
									r, err := api.CreateRelease(context.Background(), &proto.CreateReleaseRequest{
										ID:       id,
										CocoonID: "cocoon-123",
										URL:      "https://github.com/ncodes/cocoon-example-01",
										Language: supportedLanguages[0],
									})
									So(r, ShouldBeNil)
									So(err, ShouldNotBeNil)
									So(err.Error(), ShouldEqual, "a release with matching id already exists")
								})

								Convey(".GetRelease", func() {
									Convey("Should successfully get an existing release", func() {
										r, err := api.GetRelease(context.Background(), &proto.GetReleaseRequest{
											ID: id,
										})
										So(err, ShouldBeNil)
										So(r.Status, ShouldEqual, 200)
										So(len(r.Body), ShouldNotEqual, 0)
									})

									Convey("Should return error if release is not found", func() {
										r, err := api.GetRelease(context.Background(), &proto.GetReleaseRequest{
											ID: id,
										})
										So(err, ShouldBeNil)
										So(r.Status, ShouldEqual, 200)
										So(len(r.Body), ShouldNotEqual, 0)
									})
								})
							})
						})
					})
				})

				Convey(".GetIdentity", func() {

					identity := types.Identity{Email: util.RandString(5) + "@example.com"}
					i, err := api.CreateIdentity(context.Background(), &proto.CreateIdentityRequest{
						Email:    identity.Email,
						Password: "somepassword",
					})
					So(err, ShouldBeNil)
					So(i.Status, ShouldEqual, 200)

					Convey("Should successfully get an existing identity by email", func() {
						i, err := api.GetIdentity(context.Background(), &proto.GetIdentityRequest{
							Email: identity.Email,
						})
						So(err, ShouldBeNil)
						So(i.Status, ShouldEqual, 200)
						So(len(i.Body), ShouldNotEqual, 0)
					})

					Convey("Should successfully get an existing identity by ID", func() {
						i, err := api.GetIdentity(context.Background(), &proto.GetIdentityRequest{
							ID: identity.GetID(),
						})
						So(err, ShouldBeNil)
						So(i.Status, ShouldEqual, 200)
						So(len(i.Body), ShouldNotEqual, 0)
					})
				})

				Convey(".Login", func() {
					email := util.RandString(5) + "@example.com"
					i, err := api.CreateIdentity(context.Background(), &proto.CreateIdentityRequest{
						Email:    email,
						Password: "somepassword",
					})
					So(err, ShouldBeNil)
					So(i.Status, ShouldEqual, 200)

					Convey("Should successfully authenticate an identity", func() {
						r, err := api.Login(context.Background(), &proto.LoginRequest{
							Email:    email,
							Password: "somepassword",
						})
						So(err, ShouldBeNil)
						So(r.Status, ShouldEqual, 200)
					})

					Convey("Should return error if identity email is invalid", func() {
						r, err := api.Login(context.Background(), &proto.LoginRequest{
							Email:    "invalid@example.com",
							Password: "somepassword",
						})
						So(r, ShouldBeNil)
						So(err, ShouldNotBeNil)
						So(err.Error(), ShouldEqual, "email or password are invalid")
					})

					Convey("Should return error if identity password is invalid", func() {
						r, err := api.Login(context.Background(), &proto.LoginRequest{
							Email:    email,
							Password: "somewrongpassword",
						})
						So(r, ShouldBeNil)
						So(err, ShouldNotBeNil)
						So(err.Error(), ShouldEqual, "email or password are invalid")
					})
				})
			})

			close(apiEndCh)
		})

		close(endCh)
		err := dropDB(t)
		if err != nil {
			t.Fail()
		}
	})
}
