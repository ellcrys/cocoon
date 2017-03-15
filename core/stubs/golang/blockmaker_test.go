package golang

import (
	"testing"
	"time"

	"github.com/ncodes/cocoon/core/types"
	"github.com/ncodes/cocoon/core/types"
	. "github.com/smartystreets/goconvey/convey"
)

func TestBlockMaker(t *testing.T) {
	Convey("BlockMaker", t, func() {

		bm := NewBlockMaker(10, 10*time.Millisecond)
		So(bm.entryQueue.Empty(), ShouldEqual, true)
		So(bm.t, ShouldBeNil)

		Convey(".GetBlockMaxSize", func() {
			Convey("Should return 10", func() {
				So(bm.GetBlockMaxSize(), ShouldEqual, 10)
			})
		})

		Convey(".Add", func() {
			Convey("Should successfully add some entries", func() {
				bm.Add(&Entry{Tx: &types.Transaction{Number: 1}, RespChan: make(chan interface{})})
				bm.Add(&Entry{Tx: &types.Transaction{Number: 1}, RespChan: make(chan interface{})})
				So(bm.entryQueue.Size(), ShouldEqual, 2)
			})
		})

		Convey(".getBlockEntries", func() {
			Convey("Should return all entries in the queue", func() {
				bm.Add(&Entry{Tx: &types.Transaction{Number: 1}, RespChan: make(chan interface{})})
				bm.Add(&Entry{Tx: &types.Transaction{Number: 1}, RespChan: make(chan interface{})})
				entries := bm.getBlockEntries()
				So(len(entries), ShouldEqual, 2)
				So(bm.entryQueue.Empty(), ShouldEqual, true)
			})
		})

		Convey(".sendToEntries", func() {
			Convey("Should successfully receive message from entries channel", func() {
				ch1 := make(chan interface{})
				ch2 := make(chan interface{})
				bm.Add(&Entry{Tx: &types.Transaction{Number: 1}, RespChan: ch1})
				bm.Add(&Entry{Tx: &types.Transaction{Number: 1}, RespChan: ch2})
				entries := bm.getBlockEntries()
				So(len(entries), ShouldEqual, 2)
				go bm.sendToEntries(entries, nil)
				So(<-ch1, ShouldBeNil)
				So(<-ch2, ShouldBeNil)

				Convey("Should be unable to send to the now closed entry channel", func() {
					So(func() {
						ch1 <- nil
						ch2 <- nil
					}, ShouldPanicWith, "send on closed channel")
				})
			})
		})

		Convey(".Begin", func() {
			Convey("Should call the committer function and pass entries if entries are available", func() {
				ch1 := make(chan interface{}, 1)
				ch2 := make(chan interface{}, 1)
				ch3 := make(chan interface{}, 1)
				bm.Add(&Entry{Tx: &types.Transaction{Ledger: "a"}, RespChan: ch1})
				bm.Add(&Entry{Tx: &types.Transaction{Ledger: "b"}, RespChan: ch2})
				bm.Add(&Entry{Tx: &types.Transaction{Ledger: "a"}, RespChan: ch3})
				block := types.Block{ID: "block1"}

				committerCallCount := 0
				go bm.Begin(func(entries []*Entry) interface{} {
					committerCallCount++
					return &block
				})

				time.Sleep(15 * time.Millisecond)
				So(<-ch1, ShouldResemble, &block)
				So(<-ch2, ShouldResemble, &block)
				So(<-ch3, ShouldResemble, &block)
				So(committerCallCount, ShouldEqual, 2)
				So(bm.entryQueue.Empty(), ShouldEqual, true)
				bm.Stop()
			})
		})

		Convey(".groupEntriesByLedgerName", func() {
			Convey("Should return expected entries in expected order", func() {
				ch1 := make(chan interface{})
				a := &Entry{Tx: &types.Transaction{Ledger: "a"}, RespChan: ch1}
				a2 := &Entry{Tx: &types.Transaction{Ledger: "a"}, RespChan: ch1}
				c := &Entry{Tx: &types.Transaction{Ledger: "c"}, RespChan: ch1}
				b := &Entry{Tx: &types.Transaction{Ledger: "b"}, RespChan: ch1}
				ab := &Entry{Tx: &types.Transaction{Ledger: "ab"}, RespChan: ch1}
				entries := []*Entry{a, a2, c, b, ab}
				grp := bm.groupEntriesByLedgerName(entries)
				expected := [][]*Entry{
					[]*Entry{a, a2},
					[]*Entry{ab},
					[]*Entry{b},
					[]*Entry{c},
				}
				So(grp, ShouldResemble, expected)
			})
		})
	})
}
