package golang

import (
	"fmt"
	"strconv"

	"github.com/ellcrys/util"
	"github.com/ncodes/cocoon/core/common"
	"github.com/ncodes/cocoon/core/stubs/golang/proto"
	"github.com/ncodes/cocoon/core/types"
)

// RangeGetter defines a means to fetch keys of
// a specific range with support for
type RangeGetter struct {
	ledgerName string
	start      string
	end        string
	limit      int
	offset     int
	txs        []*types.Transaction
	curIndex   int
}

// NewRangeGetter creates a new range getter
func NewRangeGetter(ledgerName string, start, end string) *RangeGetter {
	return &RangeGetter{
		ledgerName: ledgerName,
		start:      start,
		end:        end,
		limit:      10,
		offset:     0,
	}
}

func (rg *RangeGetter) fetch() error {

	var respCh = make(chan *proto.Tx)
	err := sendTx(&proto.Tx{
		Id:     util.UUID4(),
		Invoke: true,
		Name:   types.TxRangeGet,
		Params: []string{rg.ledgerName, rg.start, rg.end, strconv.Itoa(rg.limit), strconv.Itoa(rg.offset)},
	}, respCh)
	if err != nil {
		return fmt.Errorf("failed to get transactions. %s", err)
	}

	resp, err := common.AwaitTxChan(respCh)
	if err != nil {
		return err
	}
	if resp.Status != 200 {
		return fmt.Errorf("%s", common.StripRPCErrorPrefix(resp.Body))
	}

	var txs []*types.Transaction
	if err = util.FromJSON(resp.Body, &txs); err != nil {
		return fmt.Errorf("failed to unmarshall response data")
	}

	rg.txs = append(rg.txs, txs...)
	return nil
}

// HasRow determines whether more rows exists.
func (rg *RangeGetter) HasRow() (bool, error) {
	if !isConnected() {
		return false, ErrNotConnected
	}

	if len(rg.txs) == 0 {
		if err := rg.fetch(); err != nil {
			return false, err
		}
		if len(rg.txs) == 0 {
			return false, nil
		}
		rg.curIndex = 0
	}

	return true, nil
}
