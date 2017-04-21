package lock

import (
	"testing"

	"github.com/ellcrys/util"
	"github.com/ncodes/cocoon/core/types"
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

				Convey("Should have no problem re-acquiring a lock as long as TTL has not passed", func() {
					err := l.Acquire(key)
					So(err, ShouldBeNil)
				})

				Convey("Should fail if lock has already been acquired by a different session", func() {
					l := NewConsulLock()
					err := l.Acquire(key)
					So(err, ShouldResemble, types.ErrLockAlreadyAcquired)
				})
			})
		})

		Convey(".ReleaseLock", func() {
			Convey("Should successfully release an acquired lock", func() {
				key := util.RandString(10)
				l := NewConsulLock()
				err := l.Acquire(key)
				So(err, ShouldBeNil)
				err = l.Release()
				So(err, ShouldBeNil)

				Convey("Should successfully acquire a released lock", func() {
					err := l.Acquire(key)
					So(err, ShouldBeNil)
				})
			})
		})

		Convey(".IsAcquirer", func() {
			Convey("Should return nil if lock is still the acquirer of a lock on it's key", func() {
				key := util.RandString(10)
				l := NewConsulLock()
				err := l.Acquire(key)
				So(err, ShouldBeNil)
				err = l.IsAcquirer()
				So(err, ShouldBeNil)
			})
		})
	})
}
