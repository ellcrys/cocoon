package lock

import (
	"testing"

	"github.com/ellcrys/util"
	. "github.com/smartystreets/goconvey/convey"
)

func TestFunc(t *testing.T) {
	Convey("ConsulLock", t, func() {

		Convey(".AcquireLock", func() {
			Convey("Should successfully acquire a lock", func() {
				key := util.RandString(10)
				l := NewConsulLock()
				err := l.Acquire(key)
				So(err, ShouldBeNil)

				Convey("Should have no problem calling re-acquiring a lock as long as TTL has not passed", func() {
					err := l.Acquire(key)
					So(err, ShouldBeNil)
				})
			})
		})
	})
}
