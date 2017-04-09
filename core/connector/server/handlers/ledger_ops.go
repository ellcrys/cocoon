package handlers

import (
	"fmt"
	"strconv"

	"github.com/ellcrys/util"
	"github.com/ncodes/cocoon/core/connector"
	"github.com/ncodes/cocoon/core/connector/server/acl"
	"github.com/ncodes/cocoon/core/connector/server/connector_proto"
	"github.com/ncodes/cocoon/core/orderer"
	orderer_proto "github.com/ncodes/cocoon/core/orderer/proto"
	"github.com/ncodes/cocoon/core/types"
	context "golang.org/x/net/context"
)

// LedgerOperations represents a ledger operation handler
type LedgerOperations struct {
	CocoonID         string
	ordererDiscovery *orderer.Discovery
	connector        *connector.Connector
}

// NewLedgerOperationHandler creates a new instance of a ledger operation handler
func NewLedgerOperationHandler(connector *connector.Connector, ordererDiscovery *orderer.Discovery) *LedgerOperations {
	return &LedgerOperations{
		connector:        connector,
		CocoonID:         connector.GetRequest().ID,
		ordererDiscovery: ordererDiscovery,
	}
}

// checkACL checks the operation against the ACL rules of the access cocoon
// that owns the ledger being accessed.
func (l *LedgerOperations) checkACL(op *connector_proto.LedgerOperation) error {

	ledgerName := op.GetParams()[0]

	// link is not a principal link (principal links access their own resources).
	// and the cocoon being accessed in the system cocoon, apply system ACL rules.
	if op.GetLinkTo() != l.CocoonID && op.GetLinkTo() == types.SystemCocoonID {
		i := acl.NewInterpreter(acl.SystemACL, ledgerName == "public")
		if errs := i.Validate(); len(errs) != 0 {
			return fmt.Errorf("system acl is not valid")
		}
		if !i.IsAllowed(ledgerName, l.CocoonID, op.Name) {
			return fmt.Errorf("permission denied: %s operation not allowed", op.Name)
		}
	}

	// TODO: Apply ACL rules for links pointing to non-system cocoons

	return nil
}

// Handle handles all types of ledger operations
func (l *LedgerOperations) Handle(ctx context.Context, op *connector_proto.LedgerOperation) (*connector_proto.Response, error) {

	if err := l.checkACL(op); err != nil {
		return nil, err
	}

	switch op.GetName() {
	case types.TxCreateLedger:
		return l.createLedger(ctx, op)
	case types.TxGetLedger:
		return l.getLedger(ctx, op)
	case types.TxPut:
		return l.put(ctx, op)
	case types.TxGet:
		return l.get(ctx, op, false)
	case types.TxGetByID:
		return l.get(ctx, op, true)
	case types.TxGetBlockByID:
		return l.getBlock(ctx, op)
	case types.TxRangeGet:
		return l.getRange(ctx, op)
	default:
		return nil, fmt.Errorf("unsupported operation [%s]", op.GetName())
	}
}

// createLedger sends a request to the orderer
// to create a new ledger.
func (l *LedgerOperations) createLedger(ctx context.Context, op *connector_proto.LedgerOperation) (*connector_proto.Response, error) {

	var cocoonID = l.CocoonID
	if len(op.GetLinkTo()) > 0 {
		cocoonID = op.GetLinkTo()
	}

	ordererConn, err := l.ordererDiscovery.GetGRPConn()
	if err != nil {
		return nil, err
	}
	defer ordererConn.Close()

	odc := orderer_proto.NewOrdererClient(ordererConn)
	result, err := odc.CreateLedger(ctx, &orderer_proto.CreateLedgerParams{
		CocoonID: cocoonID,
		Name:     op.GetParams()[0],
		Chained:  op.GetParams()[1] == "true",
		Public:   op.GetParams()[2] == "true",
	})

	if err != nil {
		return nil, err
	}

	body, _ := util.ToJSON(result)

	return &connector_proto.Response{
		ID:     op.GetID(),
		Status: 200,
		Body:   body,
	}, nil
}

// getLedgerOnly fetches a ledger
func (l *LedgerOperations) getLedgerOnly(ctx context.Context, cocoonID, ledgerName string) (*orderer_proto.Ledger, error) {

	ordererConn, err := l.ordererDiscovery.GetGRPConn()
	if err != nil {
		return nil, err
	}
	defer ordererConn.Close()

	odc := orderer_proto.NewOrdererClient(ordererConn)
	result, err := odc.GetLedger(ctx, &orderer_proto.GetLedgerParams{
		Name:     ledgerName,
		CocoonID: cocoonID,
	})
	if err != nil {
		return nil, err
	}

	return result, nil
}

// getLedger fetches a ledger
func (l *LedgerOperations) getLedger(ctx context.Context, op *connector_proto.LedgerOperation) (*connector_proto.Response, error) {

	var cocoonID = l.CocoonID
	if len(op.GetLinkTo()) > 0 {
		cocoonID = op.GetLinkTo()
	}

	ledger, err := l.getLedgerOnly(ctx, cocoonID, op.GetParams()[0])
	if err != nil {
		return nil, err
	}

	body, _ := util.ToJSON(ledger)

	return &connector_proto.Response{
		ID:     op.GetID(),
		Status: 200,
		Body:   body,
	}, nil
}

// put adds a new transaction to a ledger
func (l *LedgerOperations) put(ctx context.Context, op *connector_proto.LedgerOperation) (*connector_proto.Response, error) {

	var cocoonID = l.CocoonID
	if len(op.GetLinkTo()) > 0 {
		cocoonID = op.GetLinkTo()
	}

	if _, err := l.getLedger(ctx, op); err != nil {
		return nil, err
	}

	ordererConn, err := l.ordererDiscovery.GetGRPConn()
	if err != nil {
		return nil, err
	}
	defer ordererConn.Close()

	var txs []*orderer_proto.Transaction
	err = util.FromJSON(op.GetBody(), &txs)
	if err != nil {
		return nil, fmt.Errorf("failed to coerce transaction from bytes to order_proto.Transaction")
	}

	odc := orderer_proto.NewOrdererClient(ordererConn)
	result, err := odc.Put(ctx, &orderer_proto.PutTransactionParams{
		CocoonID:     cocoonID,
		LedgerName:   op.GetParams()[0],
		Transactions: txs,
	})

	if err != nil {
		return nil, err
	}

	body, _ := util.ToJSON(result.Block)

	return &connector_proto.Response{
		ID:     op.GetID(),
		Status: 200,
		Body:   body,
	}, nil
}

// get gets a transaction by its key.
// If byID is set, it will find the transaction by id specified in tx.Id.
func (l *LedgerOperations) get(ctx context.Context, op *connector_proto.LedgerOperation, byID bool) (*connector_proto.Response, error) {

	var result *orderer_proto.Transaction
	var err error

	var cocoonID = l.CocoonID
	if len(op.GetLinkTo()) > 0 {
		cocoonID = op.GetLinkTo()
	}

	ordererConn, err := l.ordererDiscovery.GetGRPConn()
	if err != nil {
		return nil, err
	}
	defer ordererConn.Close()

	odc := orderer_proto.NewOrdererClient(ordererConn)

	if !byID {
		result, err = odc.Get(ctx, &orderer_proto.GetParams{
			CocoonID: cocoonID,
			Ledger:   op.GetParams()[0],
			Key:      op.GetParams()[1],
		})
	} else {
		result, err = odc.GetByID(ctx, &orderer_proto.GetParams{
			CocoonID: cocoonID,
			Ledger:   op.GetParams()[0],
			Id:       op.GetParams()[1],
		})
	}

	if err != nil {
		return nil, err
	}

	body, _ := util.ToJSON(result)

	return &connector_proto.Response{
		ID:     op.GetID(),
		Status: 200,
		Body:   body,
	}, nil
}

// getBlock gets a block by its ledger name and id
func (l *LedgerOperations) getBlock(ctx context.Context, op *connector_proto.LedgerOperation) (*connector_proto.Response, error) {

	var cocoonID = l.CocoonID
	if len(op.GetLinkTo()) > 0 {
		cocoonID = op.GetLinkTo()
	}

	ordererConn, err := l.ordererDiscovery.GetGRPConn()
	if err != nil {
		return nil, err
	}
	defer ordererConn.Close()

	odc := orderer_proto.NewOrdererClient(ordererConn)
	block, err := odc.GetBlockByID(ctx, &orderer_proto.GetBlockParams{
		CocoonID: cocoonID,
		Ledger:   op.GetParams()[0],
		Id:       op.GetParams()[1],
	})

	if err != nil {
		return nil, err
	}

	body, _ := util.ToJSON(block)

	return &connector_proto.Response{
		ID:     op.GetID(),
		Status: 200,
		Body:   body,
	}, nil
}

// getRange fetches transactions with keys between a specified range.
func (l *LedgerOperations) getRange(ctx context.Context, op *connector_proto.LedgerOperation) (*connector_proto.Response, error) {

	var cocoonID = l.CocoonID
	if len(op.GetLinkTo()) > 0 {
		cocoonID = op.GetLinkTo()
	}

	ordererConn, err := l.ordererDiscovery.GetGRPConn()
	if err != nil {
		return nil, err
	}
	defer ordererConn.Close()

	limit, _ := strconv.Atoi(op.GetParams()[4])
	offset, _ := strconv.Atoi(op.GetParams()[5])

	odc := orderer_proto.NewOrdererClient(ordererConn)
	txs, err := odc.GetRange(ctx, &orderer_proto.GetRangeParams{
		CocoonID:  cocoonID,
		Ledger:    op.GetParams()[0],
		StartKey:  op.GetParams()[1],
		EndKey:    op.GetParams()[2],
		Inclusive: op.GetParams()[3] == "true",
		Limit:     int32(limit),
		Offset:    int32(offset),
	})

	if err != nil {
		return nil, err
	}

	body, _ := util.ToJSON(txs.Transactions)

	return &connector_proto.Response{
		ID:     op.GetID(),
		Status: 200,
		Body:   body,
	}, nil
}
