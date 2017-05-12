package archiver

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"golang.org/x/net/context"
)

func TestGStoragePersister(t *testing.T) {
	Convey("GStoragePersister", t, func() {

		Convey(".NewGStoragePersister", func() {

			Convey("Should return error if project id is not set", func() {
				_, err := NewGStoragePersister("object_name")
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "project id not set")
			})

			Convey("Should return error if object bucket name is not set", func() {
				PersisterProjectID = "some_id"
				_, err := NewGStoragePersister("object_name")
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "bucket name not set")
			})

			Convey("Should successful create an instance", func() {
				PersisterProjectID = "some_id"
				BucketName = "some_bucket"
				_, err := NewGStoragePersister("object_name")
				So(err, ShouldBeNil)
			})
		})

		Convey(".Write", func() {
			Convey("Should successful write", func() {
				PersisterProjectID = "some_id"
				BucketName = "some_bucket"
				p, err := NewGStoragePersister("object_name")
				So(err, ShouldBeNil)
				err = p.Write(context.Background(), []byte("hello"))
				So(err, ShouldBeNil)
			})
		})
	})
}
