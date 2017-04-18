package golang

import (
	"fmt"
	"time"

	"github.com/ellcrys/util"
	"github.com/ncodes/cocoon/core/common"
	"github.com/ncodes/cocoon/core/connector/server/proto_connector"
	"github.com/ncodes/cocoon/core/types"
)

// Link provides access to external services
// like ledgers, cocoon codes, event and messaging services.
type Link struct {
	cocoonID      string
	defaultLedger string
	native        bool
}

// NewLink creates a new link to a cocoon
func NewLink(cocoonID string) *Link {
	return &Link{
		cocoonID: cocoonID,
	}
}

// NewNativeLink create a new native link to a cocoon
func newNativeLink(cocoonID string) *Link {
	return &Link{
		cocoonID: cocoonID,
		native:   true,
	}
}

// IsNative checks whether the link is a native link
func (link *Link) IsNative() bool {
	return link.native
}

// SetDefaultLedger sets this link's default ledger
func (link *Link) SetDefaultLedger(name string) *Link {
	link.defaultLedger = name
	return link
}

// GetDefaultLedger returns the link's default ledger
func (link *Link) GetDefaultLedger() string {
	return link.defaultLedger
}

// GetCocoonID returns the cocoon id attached to this link
func (link *Link) GetCocoonID() string {
	return link.cocoonID
}

// NewRangeGetter creates an instance of a RangeGetter bounded to the default ledger
// and cocoon id of this link.
func (link *Link) NewRangeGetter(start, end string, inclusive bool) *RangeGetter {
	return NewRangeGetter(link.GetDefaultLedger(), link.GetCocoonID(), start, end, inclusive)
}

// NewRangeGetterFrom creates an instance of a RangeGetter for a specified ledger.
func (link *Link) NewRangeGetterFrom(ledgerName, start, end string, inclusive bool) *RangeGetter {
	return NewRangeGetter(ledgerName, link.GetCocoonID(), start, end, inclusive)
}

// CreateLedger creates a new ledger by sending an
// invoke transaction (TxCreateLedger) to the connector.
// If chained is set to true, a blockchain is created and subsequent
// PUT operations to the ledger will be included in the types. Otherwise,
// PUT operations will only be included in the types.
func (link *Link) CreateLedger(name string, chained, public bool) (*types.Ledger, error) {

	if !common.IsValidResName(name) {
		return nil, types.ErrInvalidResourceName
	}

	result, err := sendLedgerOp(&proto_connector.LedgerOperation{
		ID:     util.UUID4(),
		Name:   types.TxCreateLedger,
		LinkTo: link.GetCocoonID(),
		Params: []string{name, fmt.Sprintf("%t", chained), fmt.Sprintf("%t", public)},
	})

	if err != nil {
		return nil, err
	}

	var ledger types.Ledger
	if err = util.FromJSON(result, &ledger); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response data")
	}

	return &ledger, nil
}

// GetLedger fetches a ledger
func (link *Link) GetLedger(ledgerName string) (*types.Ledger, error) {

	result, err := sendLedgerOp(&proto_connector.LedgerOperation{
		ID:     util.UUID4(),
		Name:   types.TxGetLedger,
		LinkTo: link.GetCocoonID(),
		Params: []string{ledgerName},
	})

	if err != nil {
		return nil, err
	}

	var ledger types.Ledger
	if err = util.FromJSON(result, &ledger); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response data")
	}

	return &ledger, nil
}

// PutIn adds a new transaction to a ledger
func (link *Link) PutIn(ledgerName string, key string, value []byte) (*types.Transaction, error) {

	start := time.Now()

	if !common.IsValidResName(key) {
		return nil, types.ErrInvalidResourceName
	}

	ledger, err := link.GetLedger(ledgerName)
	if err != nil {
		return nil, err
	}

	tx := &types.Transaction{
		ID:             util.UUID4(),
		Ledger:         ledger.Name,
		LedgerInternal: types.MakeLedgerName(link.GetCocoonID(), ledger.Name),
		Key:            key,
		KeyInternal:    types.MakeTxKey(link.GetCocoonID(), key),
		Value:          string(value),
		CreatedAt:      time.Now().Unix(),
	}

	tx.Hash = tx.MakeHash()

	if ledger.Chained {
		respChan := make(chan interface{})
		blockMaker.Add(&Entry{
			Tx:       tx,
			RespChan: respChan,
			LinkTo:   link.GetCocoonID(),
		})
		result := <-respChan

		log.Debug("Put(): Time taken: ", time.Since(start))

		switch v := result.(type) {
		case error:
			return nil, v
		case *types.Block:
			tx.Block = v
			return tx, err
		default:
			return nil, fmt.Errorf("unexpected response %s", err)
		}
	}

	txJSON, _ := util.ToJSON([]*types.Transaction{tx})
	_, err = sendLedgerOp(&proto_connector.LedgerOperation{
		ID:     util.UUID4(),
		Name:   types.TxPut,
		LinkTo: link.GetCocoonID(),
		Params: []string{ledgerName},
		Body:   txJSON,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to put transaction: %s", err)
	}

	log.Debug("Put(): Time taken: ", time.Since(start))

	return tx, nil
}

// Put adds a new transaction into the default ledger
func (link *Link) Put(key string, value []byte) (*types.Transaction, error) {
	if link.GetDefaultLedger() == "" {
		return nil, fmt.Errorf("default ledger not set")
	}
	return link.PutIn(link.GetDefaultLedger(), key, value)
}

// GetFrom returns a transaction by its key and the ledger it belongs to
func (link *Link) GetFrom(ledgerName, key string) (*types.Transaction, error) {

	result, err := sendLedgerOp(&proto_connector.LedgerOperation{
		ID:     util.UUID4(),
		Name:   types.TxGet,
		LinkTo: link.GetCocoonID(),
		Params: []string{ledgerName, key},
	})

	if err != nil {
		return nil, err
	}

	var tx types.Transaction
	if err = util.FromJSON(result, &tx); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response data")
	}

	if tx.Block.ID == "" {
		tx.Block = nil
	}

	return &tx, nil
}

// Get returns a transaction that belongs to the default legder by its key.
func (link *Link) Get(key string) (*types.Transaction, error) {
	if link.GetDefaultLedger() == "" {
		return nil, fmt.Errorf("default ledger not set")
	}
	return link.GetFrom(link.GetDefaultLedger(), key)
}

// GetByIDFrom returns a transaction by its id and the ledger it belongs to
func (link *Link) GetByIDFrom(ledgerName, id string) (*types.Transaction, error) {

	result, err := sendLedgerOp(&proto_connector.LedgerOperation{
		ID:     util.UUID4(),
		Name:   types.TxGetByID,
		LinkTo: link.GetCocoonID(),
		Params: []string{ledgerName, id},
	})

	if err != nil {
		return nil, err
	}

	var tx types.Transaction
	if err = util.FromJSON(result, &tx); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response data")
	}

	if tx.Block.ID == "" {
		tx.Block = nil
	}

	return &tx, nil
}

// GetByID returns a transaction that belongs to the default legder by its id.
func (link *Link) GetByID(id string) (*types.Transaction, error) {
	if link.GetDefaultLedger() == "" {
		return nil, fmt.Errorf("default ledger not set")
	}
	return link.GetByIDFrom(link.GetDefaultLedger(), id)
}

// GetBlockFrom returns a block from a ledger by its block id
func (link *Link) GetBlockFrom(ledgerName, id string) (*types.Block, error) {

	result, err := sendLedgerOp(&proto_connector.LedgerOperation{
		ID:     util.UUID4(),
		Name:   types.TxGetBlockByID,
		LinkTo: link.GetCocoonID(),
		Params: []string{ledgerName, id},
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get block: %s", err)
	}

	var blk types.Block
	if err = util.FromJSON(result, &blk); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response data")
	}

	return &blk, nil
}

// GetBlock returns a block from the default ledger by its block id
func (link *Link) GetBlock(id string) (*types.Block, error) {
	if link.GetDefaultLedger() == "" {
		return nil, fmt.Errorf("default ledger not set")
	}
	return link.GetBlockFrom(link.GetDefaultLedger(), id)
}
