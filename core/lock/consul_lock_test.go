package lock

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestFunc(t *testing.T) {
	Convey("ConsulLock", t, func() {
		Convey(".AcquireLock", func() {
			l := NewConsulLock()
			err := l.Acquire("some_key")
			So(err, ShouldBeNil)
		})
	})
}
