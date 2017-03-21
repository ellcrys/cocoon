package golang

import (
	"sort"
	"time"

	"fmt"

	"github.com/ncodes/cocoon/core/types"
	logging "github.com/op/go-logging"
	"gopkg.in/oleiade/lane.v1"
)

var logBlockMaker = logging.MustGetLogger("ccode.stub.blockmaker")

// Entry represents a transaction to be included in the blockchain
// and a response channel to send the block information
type Entry struct {
	Tx       *types.Transaction
	To       string
	RespChan chan interface{}
}

// Entries represents a collection of entries
type Entries []*Entry

// Len returns the number of entries
func (e Entries) Len() int {
	return len(e)
}

// Less checks whether i is less than j
func (e Entries) Less(i, j int) bool {
	return fmt.Sprintf("%s.%s", e[i].To, e[i].Tx.Ledger) < fmt.Sprintf("%s.%s", e[j].To, e[j].Tx.Ledger)
}

// Swap swaps i and j
func (e Entries) Swap(i, j int) {
	e[i], e[j] = e[j], e[i]
}

// BlockMaker defines functionality for a process
// that collects transactions and periodically creates
// groups of transactions (a.k.a block), calls a committer method
// to send the block to the orderer and sends the new block to the transactions
// response channel.
type BlockMaker struct {
	entryQueue   *lane.Queue
	blockMaxSize int
	stop         chan struct{}
	t            *time.Ticker
	interval     time.Duration
}

// NewBlockMaker creates a new block maker
func NewBlockMaker(blockMaxSize int, i time.Duration) *BlockMaker {
	return &BlockMaker{
		entryQueue:   lane.NewQueue(),
		blockMaxSize: blockMaxSize,
		interval:     i,
		stop:         make(chan struct{}),
	}
}

// Add adds a transaction entry
func (b *BlockMaker) Add(entry *Entry) {
	b.entryQueue.Enqueue(entry)
}

// GetBlockMaxSize returns the maximum number of transasctions per block
func (b *BlockMaker) GetBlockMaxSize() int {
	return b.blockMaxSize
}

// GetBlockEntries creates a block of entries
func (b *BlockMaker) getBlockEntries() []*Entry {
	var entries []*Entry
	for len(entries) < b.GetBlockMaxSize() && !b.entryQueue.Empty() {
		entries = append(entries, b.entryQueue.Dequeue().(*Entry))
	}
	return entries
}

// sendToEntries sends a message to the channel of all entries
// and closes the channel.
func (b *BlockMaker) sendToEntries(entries []*Entry, msg interface{}) {
	for _, entry := range entries {
		entry.RespChan <- msg
		close(entry.RespChan)
	}
}

// Begin starts a ticker that collects a limited number of transactions
// and passes it to the commit function. Because the transactions may be associated with different
// ledgers and the orderer PUT operation requires all transactions to be of same ledger,
// transactions are collected and grouped into separate sub entries and passed to the commit function.
// The commit function determines how the blocks of entries passed to it are included in the blockchain
// The return value of the committer function is passed to all entry's/transaction's response channel
// before the channel is closed.
func (b *BlockMaker) Begin(committerFunc func([]*Entry) interface{}) {
	b.t = time.NewTicker(b.interval)
	for {
		select {
		case <-b.t.C:
			entries := b.getBlockEntries()
			if len(entries) == 0 {
				continue
			}
			entriesGrp := b.groupEntriesByLedgerName(entries)
			for _, grp := range entriesGrp {
				_grp := grp
				go func() {
					result := committerFunc(_grp)
					b.sendToEntries(_grp, result)
				}()
			}
		case <-b.stop:
			b.t.Stop()
		}
	}
}

// Stop stops the block maker
func (b *BlockMaker) Stop() {
	b.t.Stop()
	close(b.stop)
}

// groupEntriesByLedgerName creates a group of entries where all entries
// in a group all have the same LedgerName and To value.
func (b *BlockMaker) groupEntriesByLedgerName(entries Entries) [][]*Entry {
	sort.Sort(entries)
	var grp = [][]*Entry{}
	var curLedgerName = ""
	var curTo = ""
	for _, entry := range entries {
		if entry.Tx.Ledger != curLedgerName || entry.To != curTo {
			curLedgerName = entry.Tx.Ledger
			curTo = entry.To
			grp = append(grp, []*Entry{})
		}
		grp[len(grp)-1] = append(grp[len(grp)-1], entry)
	}
	return grp
}
