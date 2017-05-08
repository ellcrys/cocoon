package api

import (
	"context"
	"testing"
	"time"

	"google.golang.org/grpc/metadata"

	"os"

	"fmt"

	"database/sql"

	"github.com/ellcrys/util"
	"github.com/ncodes/cocoon/core/api/api/proto_api"
	blkch_impl "github.com/ncodes/cocoon/core/blockchain/impl"
	"github.com/ncodes/cocoon/core/orderer/orderer"
	"github.com/ncodes/cocoon/core/scheduler"
	"github.com/ncodes/cocoon/core/store/impl"
	"github.com/ncodes/cocoon/core/types"
	logging "github.com/op/go-logging"

	"strings"

	_ "github.com/lib/pq"
	. "github.com/smartystreets/goconvey/convey" // convey needs this
)

var db *sql.DB
var dbName = "test_" + strings.ToLower(util.RandString(5))
var storeConStr = util.Env("STORE_CON_STR", "host=localhost user=ned dbname="+dbName+" sslmode=disable password=")
var storeConStrWithDB = util.Env("STORE_CON_STR", "host=localhost user=ned sslmode=disable password=")

func init() {
	os.Setenv("APP_ENV", "test")

	var err error
	db, err = sql.Open("postgres", storeConStrWithDB)
	if err != nil {
		panic(fmt.Errorf("failed to connector to datatabase: %s", err))
	}
}

func createDb(t *testing.T) error {
	_, err := db.Query(fmt.Sprintf("CREATE DATABASE %s;", dbName))
	return err
}

func dropDB(t *testing.T) error {
	_, err := db.Query(fmt.Sprintf("DROP DATABASE %s;", dbName))
	return err
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
	newOrderer.EventEmitter.AddListener("started", func() {
		startCB(newOrderer, endCh)
	})
	<-endCh
}

func startAPIServer(t *testing.T, startCB func(*API, chan bool)) {
	endCh := make(chan bool)
	apiServer, err := NewAPI(scheduler.NewNomad())
	if err != nil {
		t.Error(err)
		t.Fail()
		return
	}
	addr := util.Env("API_ADDRESS", "127.0.0.1:7004")
	SetLogLevel(logging.CRITICAL)
	go apiServer.Start(addr, endCh)
	apiServer.EventEmitter.AddListener("started", func() {
		startCB(apiServer, endCh)
	})
	<-endCh
}

func TestOrderer(t *testing.T) {

	defer db.Close()
	err := createDb(t)
	if err != nil {
		t.Fatal(err)
	}
	defer dropDB(t)

	key := "secret"

	startOrderer(func(od *orderer.Orderer, endCh chan bool) {
		startAPIServer(t, func(api *API, apiEndCh chan bool) {
			Convey("API", t, func() {

				Convey(".CreateIdentity", func() {

					Convey("Should successfully create an identity", func() {
						identity := types.Identity{
							Email:    util.RandString(5) + "@example.com",
							Password: "somepassword",
						}

						i, err := api.CreateIdentity(context.Background(), &proto_api.
							CreateIdentityRequest{
							Email:    identity.Email,
							Password: identity.Password,
						})
						So(err, ShouldBeNil)
						So(i.Status, ShouldEqual, 200)

						Convey("Should return error if identity with same credentials already exists", func() {
							i, err := api.CreateIdentity(context.Background(), &proto_api.
								CreateIdentityRequest{
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
							ctx = metadata.NewIncomingContext(ctx, md)

							Convey("Should successfully create a cocoon", func() {

								id := util.UUID4()
								r, err := api.CreateCocoon(ctx, &proto_api.CocoonReleasePayloadRequest{
									CocoonID:       id,
									URL:            "https://github.com/ncodes/cocoon-example-01",
									Language:       "go",
									Memory:         512,
									NumSignatories: 1,
									SigThreshold:   1,
									CPUShare:       100,
								})
								So(err, ShouldBeNil)
								So(r.Status, ShouldEqual, 200)
								So(len(r.Body), ShouldNotEqual, 0)

								Convey("Should fail to create cocoon with an already used id", func() {
									r, err := api.CreateCocoon(ctx, &proto_api.CocoonReleasePayloadRequest{
										CocoonID:       id,
										URL:            "https://github.com/ncodes/cocoon-example-01",
										Language:       "go",
										Memory:         512,
										CPUShare:       100,
										NumSignatories: 1,
										SigThreshold:   1,
									})
									So(r, ShouldBeNil)
									So(err, ShouldNotBeNil)
									So(err.Error(), ShouldEqual, "cocoon with the same id already exists")
								})

								Convey(".GetCocoon", func() {
									Convey("Should successfully get an existing cocoon by id", func() {
										c, err := api.GetCocoon(context.Background(), &proto_api.
											GetCocoonRequest{
											ID: id,
										})
										So(err, ShouldBeNil)
										So(c.Status, ShouldEqual, 200)
									})

									Convey("Should return error if cocoon does not exists", func() {
										_, err = api.GetCocoon(context.Background(), &proto_api.
											GetCocoonRequest{
											ID: "unknown",
										})
										So(err, ShouldNotBeNil)
										So(err.Error(), ShouldEqual, "cocoon not found")
									})
								})
							})
						})

						Convey(".UpdateCocoon", func() {

							ss, err := makeAuthToken(util.UUID4(), identity.GetID(), "token.cli", time.Now().AddDate(0, 1, 0).Unix(), key)
							So(err, ShouldBeNil)
							md := metadata.Pairs("access_token", ss)
							ctx := context.Background()
							ctx = metadata.NewIncomingContext(ctx, md)

							id := util.UUID4()
							r, err := api.CreateCocoon(ctx, &proto_api.CocoonReleasePayloadRequest{
								CocoonID:       id,
								URL:            "https://github.com/ncodes/cocoon-example-01",
								Language:       "go",
								Memory:         512,
								NumSignatories: 1,
								SigThreshold:   1,
								CPUShare:       100,
							})
							So(err, ShouldBeNil)
							So(r.Status, ShouldEqual, 200)

							Convey("Should return error if logged in identity does not own the cocoon", func() {
								ss, err := makeAuthToken(util.UUID4(), "some_identity", "token.cli", time.Now().AddDate(0, 1, 0).Unix(), key)
								So(err, ShouldBeNil)
								md := metadata.Pairs("access_token", ss)
								ctx := context.Background()
								ctx = metadata.NewIncomingContext(ctx, md)
								_, err = api.UpdateCocoon(ctx, &proto_api.CocoonReleasePayloadRequest{
									CocoonID: id,
								})
								So(err, ShouldNotBeNil)
								So(err.Error(), ShouldEqual, "Permission denied: You do not have permission to perform this operation")
							})

							Convey("Should return error if a field is missing", func() {
								_, err = api.UpdateCocoon(ctx, &proto_api.CocoonReleasePayloadRequest{
									CocoonID: id,
								})
								So(err, ShouldNotBeNil)
								So(err.Error(), ShouldEqual, "resources.memory: memory is required")
							})
						})
					})
				})

				Convey(".GetIdentity", func() {

					identity := types.Identity{Email: util.RandString(5) + "@example.com"}
					i, err := api.CreateIdentity(context.Background(), &proto_api.
						CreateIdentityRequest{
						Email:    identity.Email,
						Password: "somepassword",
					})
					So(err, ShouldBeNil)
					So(i.Status, ShouldEqual, 200)

					Convey("Should successfully get an existing identity by email", func() {
						i, err := api.GetIdentity(context.Background(), &proto_api.
							GetIdentityRequest{
							Email: identity.Email,
						})
						So(err, ShouldBeNil)
						So(i.Status, ShouldEqual, 200)
						So(len(i.Body), ShouldNotEqual, 0)
					})

					Convey("Should successfully get an existing identity by ID", func() {
						i, err := api.GetIdentity(context.Background(), &proto_api.
							GetIdentityRequest{
							ID: identity.GetID(),
						})
						So(err, ShouldBeNil)
						So(i.Status, ShouldEqual, 200)
						So(len(i.Body), ShouldNotEqual, 0)
					})
				})

				Convey(".Login", func() {
					email := util.RandString(5) + "@example.com"
					i, err := api.CreateIdentity(context.Background(), &proto_api.
						CreateIdentityRequest{
						Email:    email,
						Password: "somepassword",
					})
					So(err, ShouldBeNil)
					So(i.Status, ShouldEqual, 200)

					Convey("Should successfully authenticate an identity", func() {
						r, err := api.Login(context.Background(), &proto_api.
							LoginRequest{
							Email:    email,
							Password: "somepassword",
						})
						So(err, ShouldBeNil)
						So(r.Status, ShouldEqual, 200)
					})

					Convey("Should return error if identity email is invalid", func() {
						r, err := api.Login(context.Background(), &proto_api.
							LoginRequest{
							Email:    "invalid@example.com",
							Password: "somepassword",
						})
						So(r, ShouldBeNil)
						So(err, ShouldNotBeNil)
						So(err.Error(), ShouldEqual, "email or password are invalid")
					})

					Convey("Should return error if identity password is invalid", func() {
						r, err := api.Login(context.Background(), &proto_api.
							LoginRequest{
							Email:    email,
							Password: "somewrongpassword",
						})
						So(r, ShouldBeNil)
						So(err, ShouldNotBeNil)
						So(err.Error(), ShouldEqual, "email or password are invalid")
					})
				})

				Convey(".CreateCocoon", func() {

				})
			})
			close(apiEndCh)
		})
		close(endCh)
	})
}
