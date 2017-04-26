package golang

import (
	"sort"
	"time"

	"fmt"

	"github.com/ncodes/cocoon/core/types"
	logging "github.com/op/go-logging"
	"gopkg.in/oleiade/lane.v1"
)

var logblockMaker = logging.MustGetLogger("ccode.stub.blockmaker")

// entry represents a transaction to be included in the blockchain
// and a response channel to send the block information
type entry struct {
	Tx       *types.Transaction
	LinkTo   string
	RespChan chan interface{}
}

// entries represents a collection of entries
type entries []*entry

// Len returns the number of entries
func (e entries) Len() int {
	return len(e)
}

// Less checks whether i is less than j
func (e entries) Less(i, j int) bool {
	// Compare using a combination of the link and ledger name
	return fmt.Sprintf("%s.%s", e[i].LinkTo, e[i].Tx.Ledger) < fmt.Sprintf("%s.%s", e[j].LinkTo, e[j].Tx.Ledger)
}

// Swap swaps i and j
func (e entries) Swap(i, j int) {
	e[i], e[j] = e[j], e[i]
}

// blockMaker defines functionality for a process
// that collects transactions and periodically creates
// groups of transactions (a.k.a block), calls a committer method
// to send the block to the orderer and sends the new block to the transactions
// response channel.
type blockMaker struct {
	entryQueue   *lane.Queue
	blockMaxSize int
	stop         chan struct{}
	t            *time.Ticker
	interval     time.Duration
}

// newblockMaker creates a new block maker
func newblockMaker(blockMaxSize int, i time.Duration) *blockMaker {
	return &blockMaker{
		entryQueue:   lane.NewQueue(),
		blockMaxSize: blockMaxSize,
		interval:     i,
		stop:         make(chan struct{}),
	}
}

// Add adds a transaction entry
func (b *blockMaker) Add(entry *entry) {
	b.entryQueue.Enqueue(entry)
}

// GetBlockMaxSize returns the maximum number of transactions per block
func (b *blockMaker) GetBlockMaxSize() int {
	return b.blockMaxSize
}

// GetBlockentries creates a block of entries
func (b *blockMaker) getBlockentries() []*entry {
	var entries []*entry
	for len(entries) < b.GetBlockMaxSize() && !b.entryQueue.Empty() {
		entries = append(entries, b.entryQueue.Dequeue().(*entry))
	}
	return entries
}

// sendToEntries sends a message to the channel of all entries
// and closes the channel. If the message is a PutResult, it finds the
// transaction receipt of the entry and sends an error to the channel if the
// entry's transaction has an err. Otherwise, it sends the block.
func (b *blockMaker) sendToEntries(_entries []*entry, msg interface{}) {
	for _, _entry := range _entries {
		switch v := msg.(type) {
		case error:
			_entry.RespChan <- v
		case *types.PutResult:
			isErr := false
			for _, r := range v.TxReceipts {
				if r.ID == _entry.Tx.ID && len(r.Err) > 0 {
					_entry.RespChan <- fmt.Errorf(r.Err)
					isErr = true
					break
				}
			}
			if !isErr {
				_entry.RespChan <- v.Block
			}
		}
		close(_entry.RespChan)
	}
}

// Begin starts a ticker that collects a limited number of transactions
// and passes it to the commit function. Because the transactions may be associated with different
// ledgers and the orderer PUT operation requires all transactions to be of same ledger,
// transactions are collected and grouped into separate sub entries and passed to the commit function.
// The commit function determines how the blocks of entries passed to it are included in the blockchain
// The return value of the committer function is passed to all entry's/transaction's response channel
// before the channel is closed.
func (b *blockMaker) Begin(committerFunc func([]*entry) interface{}) {
	b.t = time.NewTicker(b.interval)
	for {
		select {
		case <-b.t.C:
			entries := b.getBlockentries()
			if len(entries) == 0 {
				continue
			}
			entriesGrp := b.groupentriesByLedgerAndLink(entries)
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
func (b *blockMaker) Stop() {
	b.t.Stop()
	close(b.stop)
}

// groupentriesByLedgerAndLink creates a group of entries where all entries
// in a group share the same ledger name and link
func (b *blockMaker) groupentriesByLedgerAndLink(_entries entries) [][]*entry {
	sort.Sort(_entries)
	var entryGroups = [][]*entry{}
	var curLedgerName = ""
	var curLink = ""
	for _, _entry := range _entries {
		if _entry.Tx.Ledger != curLedgerName || _entry.LinkTo != curLink {
			curLedgerName = _entry.Tx.Ledger
			curLink = _entry.LinkTo
			entryGroups = append(entryGroups, []*entry{})
		}
		entryGroups[len(entryGroups)-1] = append(entryGroups[len(entryGroups)-1], _entry)
	}
	return entryGroups
}
