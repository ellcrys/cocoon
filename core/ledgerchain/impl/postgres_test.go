package impl

import (
	"testing"

	"github.com/ellcrys/util"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres" // gorm requires it
	. "github.com/smartystreets/goconvey/convey"
)

func TestPosgresLedgerChain(t *testing.T) {
	Convey("PostgresLedgerChain", t, func() {

		pgChain := new(PostgresLedgerChain)

		Convey(".Connect", func() {
			Convey("should return error when unable to connect to a postgres server", func() {
				var conStr = "host=localhost user=wrong dbname=test sslmode=disable password=abc"
				db, err := pgChain.Connect(conStr)
				So(db, ShouldBeNil)
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "failed to connect to ledgerchain backend")
			})

			Convey("should successfully connect to postgres server", func() {
				var conStr = "host=localhost user=ned dbname=cocoonchain sslmode=disable password="
				db, err := pgChain.Connect(conStr)
				So(err, ShouldBeNil)
				So(db, ShouldNotBeNil)
			})
		})

		Convey(".Init", func() {

			var conStr = "host=localhost user=ned dbname=cocoon-dev sslmode=disable password="
			db, err := pgChain.Connect(conStr)
			So(err, ShouldBeNil)

			Convey("when ledger table does not exists", func() {
				Convey("should create ledger table and create a global ledger entry", func() {

					ledgerEntryExists := db.(*gorm.DB).HasTable(LedgerListName)
					So(ledgerEntryExists, ShouldEqual, false)

					err := pgChain.Init()
					So(err, ShouldBeNil)

					ledgerEntryExists = db.(*gorm.DB).HasTable(LedgerListName)
					So(ledgerEntryExists, ShouldEqual, true)

					Convey("ledger table must include a global ledger entry", func() {
						var entries []Ledger
						err := db.(*gorm.DB).Find(&entries).Error
						So(err, ShouldBeNil)
						So(len(entries), ShouldEqual, 1)
						So(entries[0].Name, ShouldEqual, "general")
					})
				})
			})

			Convey("when ledger table exists", func() {
				Convey("should return error return nil with no effect", func() {
					err := pgChain.Init()
					So(err, ShouldBeNil)

					ledgerEntryExists := db.(*gorm.DB).HasTable(LedgerListName)
					So(ledgerEntryExists, ShouldEqual, true)

					var entries []Ledger
					err = db.(*gorm.DB).Find(&entries).Error
					So(err, ShouldBeNil)
					So(len(entries), ShouldEqual, 1)
					So(entries[0].Name, ShouldEqual, "general")
				})
			})

			Reset(func() {
				db.(*gorm.DB).DropTable(LedgerListName)
			})
		})

		Convey(".MakeLegderHash", func() {

			Convey("should return expected ledger hash", func() {
				hash := pgChain.MakeLegderHash(&Ledger{
					PrevLedgerHash: "0000000000000000000000000000000000000000000000000000000000000000",
					Name:           "general",
					CocoonCodeID:   "xh6549dh",
					Public:         true,
					CreatedAt:      1488196279,
				})
				So(hash, ShouldEqual, "fa375c76226c54885bac292cdc722017743aae83e667f7ee92e9430d112218e1")
			})

		})

		Convey(".CreateLedger", func() {

			var ledgerName = util.RandString(10)
			var conStr = "host=localhost user=ned dbname=cocoon-dev sslmode=disable password="
			db, err := pgChain.Connect(conStr)
			So(err, ShouldBeNil)
			err = pgChain.Init()
			So(err, ShouldBeNil)
			db.(*gorm.DB).LogMode(false)

			Convey("should successfully create a ledger list entry", func() {
				ledger, err := pgChain.CreateLedger(ledgerName, "cocoon_id", true)
				So(err, ShouldBeNil)
				So(ledger, ShouldNotBeNil)

				Convey("should fail since a ledger with same name already exists", func() {
					_, err := pgChain.CreateLedger(ledgerName, "cocoon_id", true)
					So(err, ShouldNotBeNil)
					So(err.Error(), ShouldEqual, `pq: duplicate key value violates unique constraint "uix_ledgers_hash"`)
				})
			})

			Reset(func() {
				db.(*gorm.DB).DropTable(LedgerListName)
			})
		})

	})
}
