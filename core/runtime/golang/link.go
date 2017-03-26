package golang

import (
	"fmt"
	"strings"
	"time"

	"github.com/ellcrys/util"
	"github.com/ncodes/cocoon/core/common"
	"github.com/ncodes/cocoon/core/runtime/golang/proto"
	"github.com/ncodes/cocoon/core/types"
)

// Link provides access to external services
// like ledgers, cocoon codes, event and messaging services.
type Link struct {
	cocoonID      string
	defaultLedger string
	native        bool
}

// NewLink creates a new instance that represents
// a link to a cocoon
func NewLink(cocoonID string) *Link {
	return &Link{
		cocoonID: cocoonID,
	}
}

// NewNativeLink create a new instance of the running cocoon or
// a natively linked cocoon.
func NewNativeLink(cocoonID string) *Link {
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

// CreateLedger creates a new ledger by sending an
// invoke transaction (TxCreateLedger) to the connector.
// If chained is set to true, a blockchain is created and subsequent
// PUT operations to the ledger will be included in the types. Otherwise,
// PUT operations will only be incuded in the types.
func (link *Link) CreateLedger(name string, chained, public bool) (*types.Ledger, error) {

	if util.Sha256(name) == GetGlobalLedger() {
		return nil, fmt.Errorf("cannot use a reserved name")
	} else if !common.IsValidResName(name) {
		return nil, fmt.Errorf("invalid ledger name")
	}

	if !isConnected() {
		return nil, ErrNotConnected
	}

	var respCh = make(chan *proto.Tx)

	txID := util.UUID4()
	err := sendTx(&proto.Tx{
		Id:     txID,
		Invoke: true,
		LinkTo: link.GetCocoonID(),
		Name:   types.TxCreateLedger,
		Params: []string{name, fmt.Sprintf("%t", chained), fmt.Sprintf("%t", public)},
	}, respCh)
	if err != nil {
		return nil, fmt.Errorf("failed to create ledger. %s", err)
	}

	resp, err := common.AwaitTxChan(respCh)
	if err != nil {
		return nil, err
	}

	if resp.Status != 200 {
		err = fmt.Errorf("%s", common.GetRPCErrDesc(fmt.Errorf("%s", resp.Body)))
		if strings.Contains(err.Error(), "already exists") {
			return nil, ErrAlreadyExist
		}
		return nil, err
	}

	var ledger types.Ledger
	if err = util.FromJSON(resp.Body, &ledger); err != nil {
		return nil, fmt.Errorf("failed to unmarshall response data")
	}

	return &ledger, nil
}

// GetLedger fetches a ledger
func (link *Link) GetLedger(ledgerName string) (*types.Ledger, error) {

	if !isConnected() {
		return nil, ErrNotConnected
	}

	var respCh = make(chan *proto.Tx)

	txID := util.UUID4()
	err := sendTx(&proto.Tx{
		Id:     txID,
		Invoke: true,
		LinkTo: link.GetCocoonID(),
		Name:   types.TxGetLedger,
		Params: []string{ledgerName},
	}, respCh)
	if err != nil {
		return nil, fmt.Errorf("failed to get ledger. %s", err)
	}

	resp, err := common.AwaitTxChan(respCh)
	if err != nil {
		return nil, err
	}
	if resp.Status != 200 {
		return nil, fmt.Errorf("%s", common.GetRPCErrDesc(fmt.Errorf("%s", resp.Body)))
	}

	var ledger types.Ledger
	if err = util.FromJSON(resp.Body, &ledger); err != nil {
		return nil, fmt.Errorf("failed to unmarshall response data")
	}

	return &ledger, nil
}

// PutIn adds a new transaction to a ledger
func (link *Link) PutIn(ledgerName string, key string, value []byte) (*types.Transaction, error) {

	start := time.Now()

	if !isConnected() {
		return nil, ErrNotConnected
	}

	ledger, err := link.GetLedger(ledgerName)
	if err != nil {
		return nil, err
	}

	tx := &types.Transaction{
		ID:        util.UUID4(),
		Ledger:    ledger.Name,
		Key:       key,
		Value:     string(value),
		CreatedAt: time.Now().Unix(),
	}

	if ledger.Chained {
		respChan := make(chan interface{})
		blockMaker.Add(&Entry{
			Tx:       tx,
			RespChan: respChan,
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

	var respCh = make(chan *proto.Tx)
	err = sendTx(&proto.Tx{
		Id:     util.UUID4(),
		Invoke: true,
		LinkTo: link.GetCocoonID(),
		Name:   types.TxPut,
		Params: []string{ledgerName},
		Body:   txJSON,
	}, respCh)

	if err != nil {
		return nil, fmt.Errorf("failed to put transaction. %s", err)
	}

	resp, err := common.AwaitTxChan(respCh)
	if err != nil {
		return nil, err
	}
	if resp.Status != 200 {
		return nil, fmt.Errorf("%s", common.GetRPCErrDesc(fmt.Errorf("%s", resp.Body)))
	}

	log.Debug("Put(): Time taken: ", time.Since(start))

	return tx, nil
}

// Put adds a new transaction into the default ledger
func (link *Link) Put(key string, value []byte) (*types.Transaction, error) {
	return link.PutIn(link.GetDefaultLedger(), key, value)
}

// GetFrom returns a transaction by its key and the ledger it belongs to
func (link *Link) GetFrom(ledgerName, key string) (*types.Transaction, error) {

	if !isConnected() {
		return nil, ErrNotConnected
	}

	var respCh = make(chan *proto.Tx)
	err := sendTx(&proto.Tx{
		Id:     util.UUID4(),
		Invoke: true,
		Name:   types.TxGet,
		LinkTo: link.GetCocoonID(),
		Params: []string{ledgerName, key},
	}, respCh)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction. %s", err)
	}

	resp, err := common.AwaitTxChan(respCh)
	if err != nil {
		return nil, err
	}
	if resp.Status != 200 {
		return nil, fmt.Errorf("%s", common.GetRPCErrDesc(fmt.Errorf("%s", resp.Body)))
	}

	var tx types.Transaction
	if err = util.FromJSON(resp.Body, &tx); err != nil {
		return nil, fmt.Errorf("failed to unmarshall response data")
	}

	return &tx, nil
}

// Get returns a transaction that belongs to the default legder by its key.
func (link *Link) Get(key string) (*types.Transaction, error) {
	return link.GetFrom(link.GetDefaultLedger(), key)
}

// GetByIDFrom returns a transaction by its id and the ledger it belongs to
func (link *Link) GetByIDFrom(ledgerName, id string) (*types.Transaction, error) {

	if !isConnected() {
		return nil, ErrNotConnected
	}

	var respCh = make(chan *proto.Tx)
	err := sendTx(&proto.Tx{
		Id:     util.UUID4(),
		Invoke: true,
		Name:   types.TxGetByID,
		LinkTo: link.GetCocoonID(),
		Params: []string{ledgerName, id},
	}, respCh)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction. %s", err)
	}

	resp, err := common.AwaitTxChan(respCh)
	if err != nil {
		return nil, err
	}
	if resp.Status != 200 {
		return nil, fmt.Errorf("%s", common.GetRPCErrDesc(fmt.Errorf("%s", resp.Body)))
	}

	var tx types.Transaction
	if err = util.FromJSON(resp.Body, &tx); err != nil {
		return nil, fmt.Errorf("failed to unmarshall response data")
	}

	return &tx, nil
}

// GetByID returns a transaction that belongs to the default legder by its id.
func (link *Link) GetByID(id string) (*types.Transaction, error) {
	return link.GetByIDFrom(link.GetDefaultLedger(), id)
}

// GetBlockFrom returns a block from a ledger by its block id
func (link *Link) GetBlockFrom(ledgerName, id string) (*types.Block, error) {

	if !isConnected() {
		return nil, ErrNotConnected
	}

	var respCh = make(chan *proto.Tx)
	err := sendTx(&proto.Tx{
		Id:     util.UUID4(),
		Invoke: true,
		LinkTo: link.GetCocoonID(),
		Name:   types.TxGetBlockByID,
		Params: []string{ledgerName, id},
	}, respCh)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction. %s", err)
	}

	resp, err := common.AwaitTxChan(respCh)
	if err != nil {
		return nil, err
	}
	if resp.Status != 200 {
		return nil, fmt.Errorf("%s", common.GetRPCErrDesc(fmt.Errorf("%s", resp.Body)))
	}

	var blk types.Block
	if err = util.FromJSON(resp.Body, &blk); err != nil {
		return nil, fmt.Errorf("failed to unmarshall response data")
	}

	return &blk, nil
}

// GetBlock returns a block from the default ledger by its block id
func (link *Link) GetBlock(id string) (*types.Block, error) {
	return link.GetBlockFrom(link.GetDefaultLedger(), id)
}
