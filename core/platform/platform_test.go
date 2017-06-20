package platform

import (
	"testing"

	"os"

	"fmt"

	"database/sql"

	"github.com/ellcrys/util"
	blkch_impl "github.com/ellcrys/cocoon/core/blockchain/impl"
	"github.com/ellcrys/cocoon/core/orderer/orderer"
	"github.com/ellcrys/cocoon/core/store/impl"
	"github.com/ellcrys/cocoon/core/types"
	logging "github.com/op/go-logging"
	"golang.org/x/net/context"

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
	os.Setenv("DEV_ORDERER_ADDR", "127.0.0.1:7013")
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

func TestPlatform(t *testing.T) {

	defer db.Close()
	err := createDb(t)
	if err != nil {
		t.Fatal(err)
	}
	defer dropDB(t)

	startOrderer(func(od *orderer.Orderer, endCh chan bool) {

		plat, err := NewPlatform()
		if err != nil {
			close(endCh)
			t.Fatalf("failed to initialize platform. %s", err)
			return
		}

		Convey("Platform", t, func() {
			Convey(".PutIdentity", func() {

				Convey("Should successfully create a new identity", func() {
					identity := &types.Identity{Email: "person@gmail.com", Password: "abcdef", ClientSessions: []string{"session_id"}}
					err := plat.PutIdentity(context.Background(), identity)
					So(err, ShouldBeNil)

					Convey(".GetIdentity", func() {

						Convey("Should successfully return an identity with private fields", func() {
							identity2, err := plat.GetIdentity(context.Background(), identity.GetID())
							So(err, ShouldBeNil)
							So(identity, ShouldResemble, identity2)
							So(identity.Password, ShouldNotBeEmpty)
							So(identity.ClientSessions, ShouldNotBeEmpty)
						})

						Convey("Should return error (ErrIdentityNotFound) if identity does not exists", func() {
							identity, err := plat.GetIdentity(context.Background(), "unknown")
							So(err, ShouldNotBeNil)
							So(err, ShouldResemble, types.ErrIdentityNotFound)
							So(identity, ShouldBeNil)
						})
					})
				})
			})

			Convey(".PutCocoon", func() {
				Convey("Should successfully create a new cocoon", func() {
					cocoon := &types.Cocoon{IdentityID: "identity_id", ID: "cocoon_id"}
					err := plat.PutCocoon(context.Background(), cocoon)
					So(err, ShouldBeNil)

					Convey(".GetCocoon", func() {
						Convey("Should successfully return a cocoon", func() {
							cocoon2, err := plat.GetCocoon(context.Background(), cocoon.ID)
							So(err, ShouldBeNil)
							So(cocoon, ShouldResemble, cocoon2)
						})

						Convey("Should return error (ErrCocoonNotFound) if cocoon does not exists", func() {
							cocoon, err := plat.GetCocoon(context.Background(), "unknown")
							So(err, ShouldNotBeNil)
							So(err, ShouldResemble, types.ErrCocoonNotFound)
							So(cocoon, ShouldBeNil)
						})
					})
				})
			})

			Convey(".PutRelease", func() {
				Convey("Should successfully create a release", func() {
					release := &types.Release{
						ID:       "release_id",
						CocoonID: "cocoon_id",
						Env: map[string]string{
							"VAR_A":         "abc",
							"VAR_B@private": "xyz",
						},
					}
					err := plat.PutRelease(context.Background(), release)
					So(err, ShouldBeNil)

					Convey(".GetRelease", func() {
						Convey("Should successfully return a release with private fields", func() {
							release2, err := plat.GetRelease(context.Background(), release.ID, true)
							So(err, ShouldBeNil)
							So(release2, ShouldResemble, release)
						})

						Convey("Should successfully return a release without private fields", func() {
							release2, err := plat.GetRelease(context.Background(), release.ID, false)
							So(err, ShouldBeNil)
							So(release2, ShouldNotResemble, release)
							So(release2.Env, ShouldHaveLength, 1)
							So(release2.Env, ShouldNotContainKey, "VAR_B@private")
						})
					})
				})
			})

			Convey(".GetCocoonAndLastActiveRelease", func() {

				cocoon := &types.Cocoon{IdentityID: "identity_id", ID: "cocoon_id2"}
				err := plat.PutCocoon(context.Background(), cocoon)
				So(err, ShouldBeNil)

				Convey("Should successfully return error if cocoon has not release and last deployed release", func() {
					_, _, err := plat.GetCocoonAndLastActiveRelease(context.Background(), cocoon.ID, false)
					So(err, ShouldNotBeNil)
					So(err.Error(), ShouldEqual, "cocoon has no release. Wierd")
				})

			})
		})

		close(endCh)
	})
}
