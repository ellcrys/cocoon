package db

import (
	"testing"

	"os"

	"github.com/boltdb/bolt"
	. "github.com/smartystreets/goconvey/convey"
)

func TestDB(t *testing.T) {

	Convey(".GetFirstByPrefix", t, func() {
		var b *bolt.Bucket
		bucket := []byte("mybucket")
		db, err := bolt.Open("/tmp/test.db", 0600, nil)
		So(err, ShouldBeNil)

		err = db.Update(func(tx *bolt.Tx) error {
			b, err = tx.CreateBucketIfNotExists(bucket)
			So(err, ShouldBeNil)
			return nil
		})
		So(err, ShouldBeNil)

		Convey("should return empty string if not found", func() {
			k, v, err := GetFirstByPrefix(db, string(bucket), "prefix")
			So(k, ShouldBeEmpty)
			So(v, ShouldBeEmpty)
			So(err, ShouldBeNil)
		})

		Convey("should return the first key with matching prefix", func() {
			err := db.Update(func(tx *bolt.Tx) error {
				b, err = tx.CreateBucketIfNotExists(bucket)
				err = b.Put([]byte("test-a"), []byte("james"))
				So(err, ShouldBeNil)
				err = b.Put([]byte("test-b"), []byte("bond"))
				So(err, ShouldBeNil)
				return nil
			})
			So(err, ShouldBeNil)

			k, v, err := GetFirstByPrefix(db, string(bucket), "test")
			So(err, ShouldBeNil)
			So(string(k), ShouldEqual, "test-a")
			So(string(v), ShouldEqual, "james")
		})

		Reset(func() {
			os.Remove("/tmp/test.db")
			db.Close()
		})
	})
}
