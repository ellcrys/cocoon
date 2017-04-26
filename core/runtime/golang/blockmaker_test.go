package golang

import (
	"testing"
	"time"

	"github.com/ncodes/cocoon/core/types"
	. "github.com/smartystreets/goconvey/convey"
)

func TestblockMaker(t *testing.T) {
	Convey("blockMaker", t, func() {

		bm := newblockMaker(10, 10*time.Millisecond)
		So(bm.entryQueue.Empty(), ShouldEqual, true)
		So(bm.t, ShouldBeNil)

		Convey(".GetBlockMaxSize", func() {
			Convey("Should return 10", func() {
				So(bm.GetBlockMaxSize(), ShouldEqual, 10)
			})
		})

		Convey(".Add", func() {
			Convey("Should successfully add some entries", func() {
				bm.Add(&entry{Tx: &types.Transaction{Number: 1}, RespChan: make(chan interface{})})
				bm.Add(&entry{Tx: &types.Transaction{Number: 1}, RespChan: make(chan interface{})})
				So(bm.entryQueue.Size(), ShouldEqual, 2)
			})
		})

		Convey(".getBlockentries", func() {
			Convey("Should return all entries in the queue", func() {
				bm.Add(&entry{Tx: &types.Transaction{Number: 1}, RespChan: make(chan interface{})})
				bm.Add(&entry{Tx: &types.Transaction{Number: 1}, RespChan: make(chan interface{})})
				entries := bm.getBlockentries()
				So(len(entries), ShouldEqual, 2)
				So(bm.entryQueue.Empty(), ShouldEqual, true)
			})
		})

		Convey(".sendToentries", func() {
			Convey("Should successfully receive message from entries channel", func() {
				ch1 := make(chan interface{})
				ch2 := make(chan interface{})
				bm.Add(&entry{Tx: &types.Transaction{Number: 1}, RespChan: ch1})
				bm.Add(&entry{Tx: &types.Transaction{Number: 1}, RespChan: ch2})
				entries := bm.getBlockentries()
				So(len(entries), ShouldEqual, 2)
				go bm.sendToentries(entries, nil)
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
				bm.Add(&entry{Tx: &types.Transaction{Ledger: "a"}, RespChan: ch1})
				bm.Add(&entry{Tx: &types.Transaction{Ledger: "b"}, RespChan: ch2})
				bm.Add(&entry{Tx: &types.Transaction{Ledger: "a"}, RespChan: ch3})
				block := types.Block{ID: "block1"}

				committerCallCount := 0
				go bm.Begin(func(entries []*entry) interface{} {
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

		Convey(".groupentriesByLedgerAndLink", func() {
			Convey("Should return expected entries in expected order", func() {
				ch1 := make(chan interface{})
				var entries entries = []*entry{}
				a := &entry{Tx: &types.Transaction{Ledger: "a"}, RespChan: ch1, LinkTo: "c1"}
				a2 := &entry{Tx: &types.Transaction{Ledger: "a"}, RespChan: ch1, LinkTo: "c1"}
				c := &entry{Tx: &types.Transaction{Ledger: "c"}, RespChan: ch1, LinkTo: "c3"}
				b := &entry{Tx: &types.Transaction{Ledger: "b"}, RespChan: ch1, LinkTo: "c2"}
				ab := &entry{Tx: &types.Transaction{Ledger: "ab"}, RespChan: ch1, LinkTo: "c4"}
				a3 := &entry{Tx: &types.Transaction{Ledger: "a"}, RespChan: ch1, LinkTo: "ab"}
				entries = append(entries, a)
				entries = append(entries, c)
				entries = append(entries, b)
				entries = append(entries, ab)
				entries = append(entries, a2)
				entries = append(entries, a3)
				grp := bm.groupentriesByLedgerAndLink(entries)
				expected := [][]*entry{
					[]*entry{a3},
					[]*entry{a, a2},
					[]*entry{b},
					[]*entry{c},
					[]*entry{ab},
				}
				So(grp, ShouldResemble, expected)
			})

		})
	})
}
