package stub

import (
	"fmt"
	"strconv"

	"github.com/ellcrys/util"
	"github.com/ellcrys/cocoon/core/connector/server/proto_connector"
	"github.com/ellcrys/cocoon/core/types"
)

var (
	// ErrNoTransaction represents a lack of transaction
	ErrNoTransaction = fmt.Errorf("no transaction")
)

// RangeGetter defines a means to fetch keys of
// a specific range with support for
type RangeGetter struct {
	to         string
	ledgerName string
	start      string
	end        string
	inclusive  bool
	limit      int
	offset     int
	txs        []*types.Transaction
	curIndex   int
	curPage    int
	Error      error
}

// NewRangeGetter creates a new range getter to traverse a specific ledger.
// Set start and end key will get transactions between those keys. However, the end key is
// not included by default. Set inclusive to true to include the end key. Set only start to
// get all keys matching that key as a prefix or set only end key to get match keys as a surfix.
func NewRangeGetter(ledgerName, to, start, end string, inclusive bool) *RangeGetter {
	return &RangeGetter{
		to:         to,
		ledgerName: ledgerName,
		start:      start,
		end:        end,
		inclusive:  inclusive,
		limit:      50,
		offset:     0,
		curPage:    0,
	}
}

// fetch transactions
func (rg *RangeGetter) fetch() error {

	result, err := sendLedgerOp(&proto_connector.LedgerOperation{
		ID:     util.UUID4(),
		Name:   types.TxRangeGet,
		LinkTo: rg.to,
		Params: []string{rg.ledgerName, rg.start, rg.end, strconv.FormatBool(rg.inclusive), strconv.Itoa(rg.limit), strconv.Itoa(rg.offset)},
	})

	if err != nil {
		return fmt.Errorf("failed to get transactions. %s", err)
	}

	var txs []*types.Transaction
	if err = util.FromJSON(result, &txs); err != nil {
		return fmt.Errorf("failed to unmarshall response data")
	}

	rg.txs = append(rg.txs, txs...)

	rg.curPage++
	rg.offset = rg.curPage * rg.limit
	return nil
}

// HasNext determines whether more rows exists.
func (rg *RangeGetter) HasNext() bool {

	if len(rg.txs) == 0 {
		if err := rg.fetch(); err != nil {
			rg.Error = err
			return false
		}
		if len(rg.txs) == 0 {
			rg.Error = ErrNoTransaction
			return false
		}
	}

	return true
}

// Next returns a transaction if available or nil
func (rg *RangeGetter) Next() *types.Transaction {
	if len(rg.txs) == 0 {
		return nil
	}

	tx := rg.txs[0]
	rg.txs = rg.txs[1:]
	return tx
}

// Reset the state for reuse
func (rg *RangeGetter) Reset() {
	rg.txs = []*types.Transaction{}
	rg.limit = 10
	rg.offset = 0
	rg.curPage = 0
	rg.curIndex = 0
	rg.Error = nil
}
