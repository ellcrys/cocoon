package golang

import (
	"fmt"
	"time"

	"github.com/ellcrys/util"
	"github.com/ncodes/cocoon/core/common"
	"github.com/ncodes/cocoon/core/connector/server/proto_connector"
	"github.com/ncodes/cocoon/core/types"
)

// ErrObjectLocked reprents an error about an object or transaction locked by another process
var ErrObjectLocked = fmt.Errorf("failed to acquire lock. object has been locked by another process")

// Metadata defines a structure for storing link related header information
type Metadata map[string]string

// Set sets a key and its value
func (m Metadata) Set(key, val string) {
	m[key] = val
}

// Get returns the value of a key
func (m Metadata) Get(key string) string {
	return m[key]
}

// Has checks whether a key exists in the object
func (m Metadata) Has(key string) bool {
	return m[key] != ""
}

// Link provides access to all platform services available to
// the cocoon code.
type Link struct {
	cocoonID   string
	native     bool
	InMetadata Metadata // incoming metadata
}

// NewLink creates a new link to a cocoon
func NewLink(cocoonID string) *Link {
	return &Link{
		cocoonID:   cocoonID,
		InMetadata: make(Metadata),
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

// GetCocoonID returns the cocoon id attached to this link
func (link *Link) GetCocoonID() string {
	return link.cocoonID
}

// GetIncomingMeta returns the metadata included in the link.
func (link *Link) GetIncomingMeta() Metadata {
	return link.InMetadata
}

// GetInvokeID returns the transaction id of the invoke request.
// Returns an empty string if link was not created as a result of an invoke request
func (link *Link) GetInvokeID() string {
	return link.InMetadata["txID"]
}

// NewRangeGetter creates an instance of a RangeGetter for a specified ledger.
func (link *Link) NewRangeGetter(ledgerName, start, end string, inclusive bool) *RangeGetter {
	return NewRangeGetter(ledgerName, link.GetCocoonID(), start, end, inclusive)
}

// NewLedger creates a new ledger by sending an
// invoke transaction (TxNewLedger) to the connector.
// If chained is set to true, a blockchain is created and subsequent
// PUT operations to the ledger will be included in the types. Otherwise,
// PUT operations will only be included in the types.
func (link *Link) NewLedger(name string, chained, public bool) (*types.Ledger, error) {

	if !common.IsValidResName(name) {
		return nil, types.ErrInvalidResourceName
	}

	result, err := sendLedgerOp(&proto_connector.LedgerOperation{
		ID:     util.UUID4(),
		Name:   types.TxNewLedger,
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

func (link *Link) put(revisionID, ledgerName, key string, value []byte) (*types.Transaction, error) {

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
		RevisionTo:     revisionID,
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
		defaultBlockMaker.Add(&entry{
			Tx:       tx,
			RespChan: respChan,
			LinkTo:   link.GetCocoonID(),
		})
		result := <-respChan

		log.Debug("Put(): Time taken: ", time.Since(start))

		switch v := result.(type) {
		case error:
			if common.CompareErr(v, ErrObjectLocked) == 0 {
				return nil, ErrObjectLocked
			}
			return nil, v
		case *types.Block:
			tx.Block = v
			return tx, err
		default:
			return nil, fmt.Errorf("unexpected response %s", err)
		}
	}

	txJSON, _ := util.ToJSON([]*types.Transaction{tx})
	putTxResultBs, err := sendLedgerOp(&proto_connector.LedgerOperation{
		ID:     util.UUID4(),
		Name:   types.TxPut,
		LinkTo: link.GetCocoonID(),
		Params: []string{ledgerName},
		Body:   txJSON,
	})

	if err != nil {
		if common.CompareErr(err, ErrObjectLocked) == 0 {
			return nil, ErrObjectLocked
		}
		return nil, fmt.Errorf("failed to put transaction: %s", err)
	}

	var putTxResult types.PutResult
	if err := util.FromJSON(putTxResultBs, &putTxResult); err != nil {
		return nil, common.JSONCoerceErr("putTxResult", err)
	}

	// ensure transaction was successful
	for _, txResult := range putTxResult.TxReceipts {
		if txResult.ID == tx.ID && len(txResult.Err) > 0 {
			return nil, fmt.Errorf("failed to put transaction: %s", txResult.Err)
		}
	}

	log.Debug("Put(): Time taken: ", time.Since(start))

	return tx, nil
}

// Put puts a transaction in a ledger
func (link *Link) Put(ledgerName, key string, value []byte) (*types.Transaction, error) {
	return link.put("", ledgerName, key, value)
}

// PutSafe puts a transaction and also ensures that no previous transaction references the revision
// id provided. If a previous transaction references the revision id, an ErrStateObject is returned.
// This is useful for situations where CAS (check-an-set) procedure is a requirement. Use the ID of
// previous transactions to ensure updates are not made based on records that may have been updated by
// another process.
func (link *Link) PutSafe(revisionID, ledgerName, key string, value []byte) (*types.Transaction, error) {
	return link.put(revisionID, ledgerName, key, value)
}

// Get gets a transaction from a ledger
func (link *Link) Get(ledgerName, key string) (*types.Transaction, error) {

	result, err := sendLedgerOp(&proto_connector.LedgerOperation{
		ID:     util.UUID4(),
		Name:   types.TxGet,
		LinkTo: link.GetCocoonID(),
		Params: []string{ledgerName, key},
	})

	if err != nil {
		if common.CompareErr(err, ErrObjectLocked) == 0 {
			return nil, ErrObjectLocked
		}
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

// GetBlock gets a block from a ledger by a block id
func (link *Link) GetBlock(ledgerName, id string) (*types.Block, error) {

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

// Lock acquires a lock on the specified key within the scope of the
// linked cocoon code. An error is returned if it failed to acquire the lock.
func (link *Link) Lock(key string, ttl time.Duration) (*Lock, error) {
	lock, err := newLock(link.cocoonID, key, ttl)
	if err != nil {
		return nil, err
	}
	if err := lock.Acquire(); err != nil {
		return nil, err
	}
	return lock, nil
}
