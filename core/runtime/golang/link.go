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
}

// NewLink creates a new instance
func NewLink() *Link {
	return &Link{}
}

// CreateLedger creates a new ledger by sending an
// invoke transaction (TxCreateLedger) to the connector.
// If chained is set to true, a blockchain is created and subsequent
// PUT operations to the ledger will be included in the types. Otherwise,
// PUT operations will only be incuded in the types.
func (link *Link) CreateLedger(name string, chained, public bool) (*types.Ledger, error) {

	if util.Sha256(name) == GetGlobalLedgerName() {
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
func GetLedger(ledgerName string) (*types.Ledger, error) {

	if !isConnected() {
		return nil, ErrNotConnected
	}

	var respCh = make(chan *proto.Tx)

	txID := util.UUID4()
	err := sendTx(&proto.Tx{
		Id:     txID,
		Invoke: true,
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
func PutIn(ledgerName string, key string, value []byte) (*types.Transaction, error) {

	start := time.Now()

	if !isConnected() {
		return nil, ErrNotConnected
	}

	ledger, err := GetLedger(ledgerName)
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
	tx.Hash = tx.MakeHash()

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
func Put(key string, value []byte) (*types.Transaction, error) {
	return PutIn(GetDefaultLedgerName(), key, value)
}

// GetFrom returns a transaction by its key and the ledger it belongs to
func GetFrom(ledgerName, key string) (*types.Transaction, error) {

	if !isConnected() {
		return nil, ErrNotConnected
	}

	var respCh = make(chan *proto.Tx)
	err := sendTx(&proto.Tx{
		Id:     util.UUID4(),
		Invoke: true,
		Name:   types.TxGet,
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
	return GetFrom(GetDefaultLedgerName(), key)
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
	return link.GetByIDFrom(GetDefaultLedgerName(), id)
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
	return link.GetBlockFrom(GetDefaultLedgerName(), id)
}
