package api

import (
	"testing"

	"golang.org/x/net/context"

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
	os.Setenv("ENV", "test")
	os.Setenv("GCP_PROJECT_ID", "visiontest-1281")

	var err error
	db, err = sql.Open("postgres", storeConStrWithDB)
	if err != nil {
		panic(fmt.Errorf("failed to connect to database: %s", err))
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

func TestRPCHandles(t *testing.T) {

	defer db.Close()
	err := createDb(t)
	if err != nil {
		t.Fatal(err)
	}
	defer dropDB(t)

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

							ctx := context.WithValue(context.Background(), types.CtxIdentity, identity.GetID())

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

							ctx := context.WithValue(context.Background(), types.CtxIdentity, identity.GetID())

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
								ctx := context.WithValue(context.Background(), types.CtxIdentity, "some_identity_id")
								_, err = api.UpdateCocoon(ctx, &proto_api.CocoonReleasePayloadRequest{
									CocoonID: id,
								})
								So(err, ShouldNotBeNil)
								So(err.Error(), ShouldEqual, "Permission denied: You do not have permission to perform this operation")
							})

							Convey("Should return no error and update nothing if no change detected", func() {
								resp, err := api.UpdateCocoon(ctx, &proto_api.CocoonReleasePayloadRequest{
									CocoonID: id,
								})
								So(err, ShouldBeNil)
								var result map[string]interface{}
								util.FromJSON(resp.Body, &result)
								So(result, ShouldContainKey, "newReleaseID")
								So(result, ShouldContainKey, "cocoonUpdated")
								So(result["newReleaseID"], ShouldBeEmpty)
								So(result["cocoonUpdated"], ShouldEqual, false)
							})

							Convey("Should return error if a field fails validation", func() {
								_, err = api.UpdateCocoon(ctx, &proto_api.CocoonReleasePayloadRequest{
									CocoonID: id,
									Language: "abc",
								})
								So(err, ShouldNotBeNil)
								So(err.Error(), ShouldContainSubstring, "language is not supported. Expects one of these ")
							})

							Convey("Should return `cocoonUpdated` set to true if cocoon field is updated because of a change", func() {
								resp, err := api.UpdateCocoon(ctx, &proto_api.CocoonReleasePayloadRequest{
									CocoonID:       id,
									NumSignatories: 2,
								})
								So(err, ShouldBeNil)
								var result map[string]interface{}
								util.FromJSON(resp.Body, &result)
								So(result, ShouldContainKey, "newReleaseID")
								So(result, ShouldContainKey, "cocoonUpdated")
								So(result["newReleaseID"], ShouldBeEmpty)
								So(result["cocoonUpdated"], ShouldEqual, true)

								cocoon, err := api.platform.GetCocoon(ctx, id)
								So(err, ShouldBeNil)
								So(cocoon.NumSignatories, ShouldEqual, 2)
							})

							Convey("Should return `newReleaseID` set if a new release is created because of a change in a release field", func() {
								resp, err := api.UpdateCocoon(ctx, &proto_api.CocoonReleasePayloadRequest{
									CocoonID: id,
									Version:  "some_version",
								})
								So(err, ShouldBeNil)
								var result map[string]interface{}
								util.FromJSON(resp.Body, &result)
								So(result, ShouldContainKey, "newReleaseID")
								So(result, ShouldContainKey, "cocoonUpdated")
								So(result["newReleaseID"], ShouldNotBeEmpty)
								So(result["cocoonUpdated"], ShouldEqual, false)

								release, err := api.platform.GetRelease(ctx, result["newReleaseID"].(string), false)
								So(err, ShouldBeNil)
								So(release.Version, ShouldEqual, "some_version")
							})
						})
						Convey(".AddSignatories", func() {

							ctx := context.WithValue(context.Background(), types.CtxIdentity, identity.GetID())
							id := util.UUID4()
							r, err := api.CreateCocoon(ctx, &proto_api.CocoonReleasePayloadRequest{
								CocoonID:       id,
								URL:            "https://github.com/ncodes/cocoon-example-01",
								Language:       "go",
								Memory:         512,
								NumSignatories: 2,
								SigThreshold:   1,
								CPUShare:       100,
							})
							So(err, ShouldBeNil)
							So(r.Status, ShouldEqual, 200)

							Convey("Should return error if identity does not exist", func() {
								resp, err := api.AddSignatories(ctx, &proto_api.AddSignatoriesRequest{
									CocoonID: id,
									IDs:      []string{"unknown_identity_id"},
								})
								So(err, ShouldBeNil)
								var result map[string]interface{}
								util.FromJSON(resp.Body, &result)
								So(result, ShouldContainKey, "added")
								So(result, ShouldContainKey, "errs")
								So(len(result["added"].([]interface{})), ShouldEqual, 0)
								So(len(result["errs"].([]interface{})), ShouldEqual, 1)
								So(result["errs"].([]interface{})[0], ShouldEqual, "unknown_ide: identity not found")
							})

							Convey("Should successfully add a signatories", func() {
								identity := &types.Identity{Email: "some@email.com"}
								err := api.platform.PutIdentity(ctx, identity)
								So(err, ShouldBeNil)
								resp, err := api.AddSignatories(ctx, &proto_api.AddSignatoriesRequest{CocoonID: id, IDs: []string{identity.GetID()}})
								So(err, ShouldBeNil)
								var result map[string]interface{}
								util.FromJSON(resp.Body, &result)
								So(result, ShouldContainKey, "added")
								So(result, ShouldContainKey, "errs")
								So(len(result["added"].([]interface{})), ShouldEqual, 1)
								So(len(result["errs"].([]interface{})), ShouldEqual, 0)
								So(result["added"].([]interface{})[0], ShouldEqual, identity.GetID())

								Convey("Should return error if maximum signatories have been added when adding a single identity", func() {
									_, err := api.AddSignatories(ctx, &proto_api.AddSignatoriesRequest{
										CocoonID: id,
										IDs:      []string{"an_id"},
									})
									So(err, ShouldNotBeNil)
									So(err.Error(), ShouldContainSubstring, "max signatories already added. You can't add more")
								})
							})

							Convey("Should return error if maximum signatories have been added when adding more than the available slots", func() {
								_, err := api.AddSignatories(ctx, &proto_api.AddSignatoriesRequest{
									CocoonID: id,
									IDs:      []string{"an_id", "an_id2"},
								})
								So(err, ShouldNotBeNil)
								So(err.Error(), ShouldContainSubstring, "maximum required signatories cannot be exceeded. You can only add 1 more signatory")
							})

							Convey("Should return error if identity is already a signatory", func() {
								err := api.platform.PutIdentity(ctx, &identity)
								So(err, ShouldBeNil)
								resp, err := api.AddSignatories(ctx, &proto_api.AddSignatoriesRequest{CocoonID: id, IDs: []string{identity.GetID()}})
								So(err, ShouldBeNil)
								var result map[string]interface{}
								util.FromJSON(resp.Body, &result)
								So(result, ShouldContainKey, "added")
								So(result, ShouldContainKey, "errs")
								So(len(result["added"].([]interface{})), ShouldEqual, 0)
								So(len(result["errs"].([]interface{})), ShouldEqual, 1)
								So(result["errs"].([]interface{})[0], ShouldContainSubstring, "identity is already a signatory")
							})

						})

						Convey(".RemoveSignatories", func() {

							ctx := context.WithValue(context.Background(), types.CtxIdentity, identity.GetID())
							id := util.UUID4()
							r, err := api.CreateCocoon(ctx, &proto_api.CocoonReleasePayloadRequest{
								CocoonID:       id,
								URL:            "https://github.com/ncodes/cocoon-example-01",
								Language:       "go",
								Memory:         512,
								NumSignatories: 2,
								SigThreshold:   1,
								CPUShare:       100,
							})
							So(err, ShouldBeNil)
							So(r.Status, ShouldEqual, 200)

							identity := &types.Identity{Email: "some@email.com"}
							err = api.platform.PutIdentity(ctx, identity)
							So(err, ShouldBeNil)
							_, err = api.AddSignatories(ctx, &proto_api.AddSignatoriesRequest{CocoonID: id, IDs: []string{identity.GetID()}})
							So(err, ShouldBeNil)

							Convey("Should successfully remove signatory", func() {
								cocoon, err := api.platform.GetCocoon(ctx, id)
								So(cocoon.Signatories, ShouldContain, identity.GetID())

								_, err = api.RemoveSignatories(ctx, &proto_api.RemoveSignatoriesRequest{CocoonID: id, IDs: []string{identity.GetID()}})
								So(err, ShouldBeNil)
								cocoon, err = api.platform.GetCocoon(ctx, id)

								So(err, ShouldBeNil)
								So(cocoon.Signatories, ShouldNotContain, identity.GetID())

								Convey("Should return no error if identity is not a signatory", func() {
									_, err = api.RemoveSignatories(ctx, &proto_api.RemoveSignatoriesRequest{CocoonID: id, IDs: []string{identity.GetID()}})
									So(err, ShouldBeNil)
								})
							})
						})

						Convey(".AddVote", func() {

							ctx := context.WithValue(context.Background(), types.CtxIdentity, identity.GetID())
							id := util.UUID4()
							r, err := api.CreateCocoon(ctx, &proto_api.CocoonReleasePayloadRequest{
								CocoonID:       id,
								URL:            "https://github.com/ncodes/cocoon-example-01",
								Language:       "go",
								Memory:         512,
								NumSignatories: 2,
								SigThreshold:   1,
								CPUShare:       100,
							})
							So(err, ShouldBeNil)
							So(r.Status, ShouldEqual, 200)
							var cocoon types.Cocoon
							util.FromJSON(r.Body, &cocoon)

							Convey("Should return error if release does not exist", func() {
								_, err := api.AddVote(ctx, &proto_api.AddVoteRequest{ReleaseID: "some_unknown_release"})
								So(err, ShouldNotBeNil)
								So(err.Error(), ShouldEqual, "release not found")
							})

							Convey("Should deny identity if identity is not a signatory", func() {
								ctx := context.WithValue(context.Background(), types.CtxIdentity, "unknown_identity")
								_, err := api.AddVote(ctx, &proto_api.AddVoteRequest{ReleaseID: cocoon.Releases[0]})
								So(err, ShouldNotBeNil)
								So(err.Error(), ShouldEqual, "Permission Denied: You are not a signatory to this cocoon")
							})

							Convey("Should successfully add a vote to a release", func() {
								_, err := api.AddVote(ctx, &proto_api.AddVoteRequest{ReleaseID: cocoon.Releases[0], Vote: 1})
								So(err, ShouldBeNil)
								release, err := api.platform.GetRelease(ctx, cocoon.Releases[0], false)
								So(err, ShouldBeNil)
								So(release.SigApproved, ShouldEqual, 1)
								So(release.SigDenied, ShouldEqual, 0)

								Convey("Should return error if identity has already added a vote", func() {
									_, err := api.AddVote(ctx, &proto_api.AddVoteRequest{ReleaseID: cocoon.Releases[0]})
									So(err, ShouldNotBeNil)
									So(err.Error(), ShouldEqual, "You have already cast a vote for this release")
								})
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

				Convey(".updateReleaseEnv", func() {

					Convey("Should successfully return updated Env", func() {
						last := types.Env(map[string]string{"VAR_A": "hello", "VAR_B": "100"})
						newest := types.Env(map[string]string{"VAR_B": "300"})
						expected := types.Env(map[string]string{"VAR_B": "300"})
						updateReleaseEnv(last, newest)
						So(newest, ShouldResemble, expected)
					})

					Convey("Should successfully include pinned variables from previous Env when the variable is included in the latest Env", func() {
						last := types.Env(map[string]string{"VAR_A@pin": "hello", "VAR_B": "100"})
						newest := types.Env(map[string]string{"VAR_A": "some_value", "VAR_B": "300"})
						expected := types.Env(map[string]string{"VAR_A@pin": "hello", "VAR_B": "300"})
						updateReleaseEnv(last, newest)
						So(newest, ShouldResemble, expected)
					})

					Convey("Should not included pinned variables from previous Env that is not in the latest Env", func() {
						last := types.Env(map[string]string{"VAR_A@pin": "hello", "VAR_B": "100"})
						newest := types.Env(map[string]string{"VAR_B": "300"})
						expected := types.Env(map[string]string{"VAR_B": "300"})
						updateReleaseEnv(last, newest)
						So(newest, ShouldResemble, expected)
					})

					Convey("Should not include pinned variables from previous Env if latest Env contains an `unpin` flag", func() {
						last := types.Env(map[string]string{"VAR_A@pin": "hello", "VAR_B": "100"})
						newest := types.Env(map[string]string{"VAR_A@unpin": "hello_world", "VAR_B": "300"})
						expected := types.Env(map[string]string{"VAR_A@unpin": "hello_world", "VAR_B": "300"})
						updateReleaseEnv(last, newest)
						So(newest, ShouldResemble, expected)
					})

					Convey("Should not include pinned variables from previous Env if latest Env contains an `unpin_once` flag but new variable flag must have `pin` flag", func() {
						last := types.Env(map[string]string{"VAR_A@pin": "hello", "VAR_B": "100"})
						newest := types.Env(map[string]string{"VAR_A@unpin_once": "hello_world", "VAR_B": "300"})
						expected := types.Env(map[string]string{"VAR_A@pin": "hello_world", "VAR_B": "300"})
						updateReleaseEnv(last, newest)
						So(newest, ShouldResemble, expected)
					})
				})
			})

			close(apiEndCh)
		})
		close(endCh)
	})
}
