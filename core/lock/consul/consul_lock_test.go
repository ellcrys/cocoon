package consul

import (
	"fmt"
	"testing"
	"time"

	"github.com/ellcrys/util"
	"github.com/ncodes/cocoon/core/types"
	. "github.com/smartystreets/goconvey/convey"
)

func TestConsulLock(t *testing.T) {
	Convey("ConsulLock", t, func() {

		Convey(".AcquireLock", func() {
			Convey("Should successfully acquire a lock", func() {
				key := util.RandString(10)
				l, err := NewLock(key)
				So(err, ShouldBeNil)
				err = l.Acquire()
				So(err, ShouldBeNil)

				Convey("Should have no problem re-acquiring a lock as long as TTL has not passed", func() {
					err := l.Acquire()
					So(err, ShouldBeNil)

					Convey("Should fail if lock has already been acquired by a different session", func() {
						l, err := NewLock(key)
						So(err, ShouldBeNil)
						err = l.Acquire()
						So(err, ShouldResemble, types.ErrLockNotAcquired)
					})
				})
			})

			Convey(".ReleaseLock", func() {
				Convey("Should successfully release an acquired lock", func() {
					key := util.RandString(10)
					l, err := NewLock(key)
					So(err, ShouldBeNil)
					err = l.Acquire()
					So(err, ShouldBeNil)
					err = l.Release()
					So(err, ShouldBeNil)

					Convey("Should successfully acquire a released lock", func() {
						err := l.Acquire()
						So(err, ShouldBeNil)
					})
				})

				Convey("Should return no error when trying to release a lock not held", func() {
					key := util.RandString(10)
					l, err := NewLock(key)
					So(err, ShouldBeNil)
					l2, err := NewLock(key)
					So(err, ShouldBeNil)
					l2.state["lock_session"] = util.UUID4()
					err = l.Acquire()
					err = l2.Release()
					So(err, ShouldBeNil)
				})
			})

			Convey(".IsAcquirer", func() {

				Convey("Should return error if lock object has no key", func() {
					l, err := NewLock("")
					So(err, ShouldBeNil)
					err = l.IsAcquirer()
					So(err, ShouldResemble, fmt.Errorf("key is not set"))
				})

				Convey("Should return nil if lock is still the acquirer of a lock on it's key", func() {
					key := util.RandString(10)
					l, err := NewLock(key)
					So(err, ShouldBeNil)
					err = l.Acquire()
					So(err, ShouldBeNil)
					err = l.IsAcquirer()
					So(err, ShouldBeNil)
				})

				Convey("Should return err if lock is no longer acquired due to TTL being reached", func() {
					key := util.RandString(10)
					l, err := NewLock(key)
					So(err, ShouldBeNil)
					err = l.Acquire()
					So(err, ShouldBeNil)
					time.Sleep(20 * time.Second)
					err = l.IsAcquirer()
					So(err, ShouldResemble, types.ErrLockNotAcquired)
				})
			})
		})
	})
}
