package impl

import (
	"testing"

	"fmt"

	"github.com/ellcrys/util"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres" // gorm requires it
	"github.com/ncodes/cocoon/core/types/blockchain"
	. "github.com/smartystreets/goconvey/convey"
)

func TestPosgresBlockchain(t *testing.T) {
	Convey("PostgresBlockchain", t, func() {

		var conStr = "host=localhost user=ned dbname=cocoon-dev sslmode=disable password="
		pgChain := new(PostgresBlockchain)
		db, err := pgChain.Connect(conStr)
		fmt.Println(err)
		So(err, ShouldBeNil)
		So(db, ShouldNotBeNil)

		var RestDB = func() {
			db.(*gorm.DB).DropTable(ChainTableName, BlockTableName)
		}

		Convey(".Connect", func() {
			Convey("should return error when unable to connect to a postgres server", func() {
				var conStr = "host=localhost user=wrong dbname=test sslmode=disable password=abc"
				pgBlkch := new(PostgresBlockchain)
				db, err := pgBlkch.Connect(conStr)
				So(db, ShouldBeNil)
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "failed to connect to blockchain backend")
			})
		})

		Convey(".Init", func() {

			Convey("when chain table does not exists", func() {

				Convey("should create chain and block table and create a global chain", func() {

					chainTableExists := db.(*gorm.DB).HasTable(ChainTableName)
					So(chainTableExists, ShouldEqual, false)

					err := pgChain.Init(blockchain.GetGlobalChainName())
					So(err, ShouldBeNil)

					chainTableExists = db.(*gorm.DB).HasTable(ChainTableName)
					So(chainTableExists, ShouldEqual, true)

					chainTableExists = db.(*gorm.DB).HasTable(ChainTableName)
					So(chainTableExists, ShouldEqual, true)

					Convey("chain table must include a global chain", func() {
						var entries []blockchain.Chain
						err := db.(*gorm.DB).Find(&entries).Error
						So(err, ShouldBeNil)
						So(len(entries), ShouldEqual, 1)
						So(entries[0].Name, ShouldEqual, blockchain.GetGlobalChainName())
					})

					Reset(func() {
						RestDB()
					})
				})
			})

			Convey("when ledger table exists", func() {
				Convey("should return nil with no effect", func() {
					err := pgChain.Init(blockchain.GetGlobalChainName())
					So(err, ShouldBeNil)

					chainTableExists := db.(*gorm.DB).HasTable(ChainTableName)
					So(chainTableExists, ShouldEqual, true)

					var chains []blockchain.Chain
					err = db.(*gorm.DB).Find(&chains).Error
					So(err, ShouldBeNil)
					So(len(chains), ShouldEqual, 1)
					So(chains[0].Name, ShouldEqual, blockchain.GetGlobalChainName())
				})

				Reset(func() {
					RestDB()
				})
			})
		})

		Convey(".MakeChainName", func() {
			Convey("Should replace namespace with empty string if provided name is equal to blockchain.GetGlobalChainName()", func() {
				name := blockchain.GetGlobalChainName()
				namespace := ""
				expected := util.Sha256(fmt.Sprintf("%s.%s", namespace, name))
				actual := pgChain.MakeChainName("namespace_will_be_ignored", name)
				So(expected, ShouldEqual, actual)
			})

			Convey("Should return expected name with namespace and name hashed together", func() {
				expected := util.Sha256(fmt.Sprintf("%s.%s", "cocooncode_1", "accounts"))
				So(expected, ShouldEqual, pgChain.MakeChainName("cocooncode_1", "accounts"))
			})
		})

		Convey(".CreateChain", func() {
			err := pgChain.Init(blockchain.GetGlobalChainName())
			So(err, ShouldBeNil)

			Convey("Should successfully create a chain", func() {
				chain, err := pgChain.CreateChain("chain1", true)
				So(err, ShouldBeNil)
				So(chain.Name, ShouldEqual, "chain1")
				So(chain.Public, ShouldEqual, true)

				Convey("Should fail when trying to create a chain with an already used name", func() {
					chain, err := pgChain.CreateChain("chain1", true)
					So(chain, ShouldBeNil)
					So(err, ShouldNotBeNil)
					So(err.Error(), ShouldEqual, "chain with matching name already exists")
				})
			})

			Reset(func() {
				RestDB()
			})
		})

		Reset(func() {
			RestDB()
		})
	})
}
