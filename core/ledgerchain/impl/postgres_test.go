package impl

import (
	"testing"

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

					ledgerEntryExists := db.(*gorm.DB).HasTable(LegderEntryName)
					So(ledgerEntryExists, ShouldEqual, false)

					err := pgChain.Init()
					So(err, ShouldBeNil)

					ledgerEntryExists = db.(*gorm.DB).HasTable(LegderEntryName)
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

					ledgerEntryExists := db.(*gorm.DB).HasTable(LegderEntryName)
					So(ledgerEntryExists, ShouldEqual, true)

					var entries []Ledger
					err = db.(*gorm.DB).Find(&entries).Error
					So(err, ShouldBeNil)
					So(len(entries), ShouldEqual, 1)
					So(entries[0].Name, ShouldEqual, "general")
				})
			})

			Reset(func() {
				db.(*gorm.DB).DropTable(LegderEntryName)
			})
		})
	})
}
