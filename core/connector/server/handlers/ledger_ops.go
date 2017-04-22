package handlers

import (
	"fmt"
	"strconv"

	"github.com/ellcrys/util"
	"github.com/ncodes/cocoon/core/common"
	"github.com/ncodes/cocoon/core/connector/connector"
	"github.com/ncodes/cocoon/core/connector/server/acl"
	"github.com/ncodes/cocoon/core/connector/server/proto_connector"
	"github.com/ncodes/cocoon/core/orderer/orderer"
	"github.com/ncodes/cocoon/core/orderer/proto_orderer"
	"github.com/ncodes/cocoon/core/types"
	logging "github.com/op/go-logging"
	context "golang.org/x/net/context"
)

// LedgerOperations represents a ledger operation handler
type LedgerOperations struct {
	CocoonID         string
	ordererDiscovery *orderer.Discovery
	connector        *connector.Connector
	log              *logging.Logger
}

// NewLedgerOperationHandler creates a new instance of a ledger operation handler
func NewLedgerOperationHandler(log *logging.Logger, connector *connector.Connector) *LedgerOperations {
	return &LedgerOperations{
		connector:        connector,
		CocoonID:         connector.GetRequest().ID,
		ordererDiscovery: connector.GetOrdererDiscoverer(),
		log:              log,
	}
}

// checkACL checks the operation against the ACL rules of the access cocoon
// that owns the ledger being accessed.
func (l *LedgerOperations) checkACL(ctx context.Context, op *proto_connector.LedgerOperation) error {

	ledgerName := op.GetParams()[0]

	// Handle link to system cocoon
	if op.LinkTo == types.SystemCocoonID {
		i := acl.NewInterpreter(acl.SystemACL, ledgerName == "public")
		if errs := i.Validate(); len(errs) != 0 {
			return fmt.Errorf("system ACL is not valid")
		}
		if !i.IsAllowed(ledgerName, l.CocoonID, op.Name) {
			return fmt.Errorf("permission denied: %s operation not allowed", op.Name)
		}
	}

	// Handle links to other cocoon code
	if op.LinkTo != l.CocoonID {

		cocoon, err := l.connector.GetCocoon(ctx, l.CocoonID)
		if err != nil {
			if common.CompareErr(err, types.ErrCocoonNotFound) == 0 {
				return fmt.Errorf("calling cocoon not found")
			}
			return err
		}

		linkedCocoon, err := l.connector.GetCocoon(ctx, op.GetLinkTo())
		if err != nil {
			if common.CompareErr(err, types.ErrCocoonNotFound) == 0 {
				return fmt.Errorf("linked cocoon not found")
			}
			return err
		}

		// Handle natively linked cocoon
		if cocoon.Link == op.LinkTo {
			// natively linked cocoon currently have same privilege as the linked cocoon
			return nil
		}

		defaultACLPolicy := false

		// get ledger and set the default ACL policy (the ledger visibility)
		if op.Name != types.TxCreateLedger {
			ledger, err := l._getLedger(ctx, op.GetLinkTo(), ledgerName)
			if err != nil {
				if common.CompareErr(err, types.ErrLedgerNotFound) == 0 {
					return fmt.Errorf("linked cocoon ledger not found")
				}
				return err
			}
			defaultACLPolicy = ledger.Public
		}

		i := acl.NewInterpreterFromACLMap(linkedCocoon.ACL, defaultACLPolicy)
		if errs := i.Validate(); len(errs) != 0 {
			return fmt.Errorf("linked cocoon ACL is not valid")
		}
		if !i.IsAllowed(ledgerName, l.CocoonID, op.Name) {
			return fmt.Errorf("permission denied: %s operation not allowed", op.Name)
		}
	}

	return nil
}

// Handle handles all types of ledger operations
func (l *LedgerOperations) Handle(ctx context.Context, op *proto_connector.LedgerOperation) (*proto_connector.Response, error) {

	if err := l.checkACL(ctx, op); err != nil {
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
func (l *LedgerOperations) createLedger(ctx context.Context, op *proto_connector.LedgerOperation) (*proto_connector.Response, error) {

	var cocoonID = l.CocoonID
	if len(op.GetLinkTo()) > 0 {
		cocoonID = op.GetLinkTo()
	}

	if !common.IsValidResName(op.GetParams()[0]) {
		return nil, types.ErrInvalidResourceName
	}

	ordererConn, err := l.ordererDiscovery.GetGRPConn()
	if err != nil {
		return nil, err
	}
	defer ordererConn.Close()

	odc := proto_orderer.NewOrdererClient(ordererConn)
	result, err := odc.CreateLedger(ctx, &proto_orderer.CreateLedgerParams{
		CocoonID: cocoonID,
		Name:     op.GetParams()[0],
		Chained:  op.GetParams()[1] == "true",
		Public:   op.GetParams()[2] == "true",
	})

	if err != nil {
		return nil, err
	}

	body, _ := util.ToJSON(result)

	return &proto_connector.Response{
		ID:     op.GetID(),
		Status: 200,
		Body:   body,
	}, nil
}

// _getLedger fetches a ledger
func (l *LedgerOperations) _getLedger(ctx context.Context, cocoonID, ledgerName string) (*proto_orderer.Ledger, error) {

	ordererConn, err := l.ordererDiscovery.GetGRPConn()
	if err != nil {
		return nil, err
	}
	defer ordererConn.Close()

	odc := proto_orderer.NewOrdererClient(ordererConn)
	result, err := odc.GetLedger(ctx, &proto_orderer.GetLedgerParams{
		Name:     ledgerName,
		CocoonID: cocoonID,
	})
	if err != nil {
		return nil, err
	}

	return result, nil
}

// getLedger fetches a ledger
func (l *LedgerOperations) getLedger(ctx context.Context, op *proto_connector.LedgerOperation) (*proto_connector.Response, error) {

	var cocoonID = l.CocoonID
	if len(op.GetLinkTo()) > 0 {
		cocoonID = op.GetLinkTo()
	}

	ledger, err := l._getLedger(ctx, cocoonID, op.GetParams()[0])
	if err != nil {
		return nil, err
	}

	body, _ := util.ToJSON(ledger)

	return &proto_connector.Response{
		ID:     op.GetID(),
		Status: 200,
		Body:   body,
	}, nil
}

// put adds a new transaction to a ledger
func (l *LedgerOperations) put(ctx context.Context, op *proto_connector.LedgerOperation) (*proto_connector.Response, error) {

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

	var txs []*proto_orderer.Transaction
	err = util.FromJSON(op.GetBody(), &txs)
	if err != nil {
		return nil, fmt.Errorf("failed to coerce transaction from bytes to order_proto.Transaction")
	}

	// validate transaction key
	for _, tx := range txs {
		if !common.IsValidResName(tx.Key) {
			return nil, fmt.Errorf("tx 0: %s", types.ErrInvalidResourceName)
		}
	}

	odc := proto_orderer.NewOrdererClient(ordererConn)
	result, err := odc.Put(ctx, &proto_orderer.PutTransactionParams{
		CocoonID:     cocoonID,
		LedgerName:   op.GetParams()[0],
		Transactions: txs,
	})

	if err != nil {
		return nil, err
	}

	body, _ := util.ToJSON(result)
	return &proto_connector.Response{
		ID:     op.GetID(),
		Status: 200,
		Body:   body,
	}, nil
}

// get gets a transaction by its key.
// If byID is set, it will find the transaction by id specified in tx.Id.
func (l *LedgerOperations) get(ctx context.Context, op *proto_connector.LedgerOperation, byID bool) (*proto_connector.Response, error) {

	var result *proto_orderer.Transaction
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

	odc := proto_orderer.NewOrdererClient(ordererConn)

	if !byID {
		result, err = odc.Get(ctx, &proto_orderer.GetParams{
			CocoonID: cocoonID,
			Ledger:   op.GetParams()[0],
			Key:      op.GetParams()[1],
		})
	} else {
		result, err = odc.GetByID(ctx, &proto_orderer.GetParams{
			CocoonID: cocoonID,
			Ledger:   op.GetParams()[0],
			Id:       op.GetParams()[1],
		})
	}

	if err != nil {
		return nil, err
	}

	body, _ := util.ToJSON(result)

	return &proto_connector.Response{
		ID:     op.GetID(),
		Status: 200,
		Body:   body,
	}, nil
}

// getBlock gets a block by its ledger name and id
func (l *LedgerOperations) getBlock(ctx context.Context, op *proto_connector.LedgerOperation) (*proto_connector.Response, error) {

	var cocoonID = l.CocoonID
	if len(op.GetLinkTo()) > 0 {
		cocoonID = op.GetLinkTo()
	}

	ordererConn, err := l.ordererDiscovery.GetGRPConn()
	if err != nil {
		return nil, err
	}
	defer ordererConn.Close()

	odc := proto_orderer.NewOrdererClient(ordererConn)
	block, err := odc.GetBlockByID(ctx, &proto_orderer.GetBlockParams{
		CocoonID: cocoonID,
		Ledger:   op.GetParams()[0],
		Id:       op.GetParams()[1],
	})

	if err != nil {
		return nil, err
	}

	body, _ := util.ToJSON(block)

	return &proto_connector.Response{
		ID:     op.GetID(),
		Status: 200,
		Body:   body,
	}, nil
}

// getRange fetches transactions with keys between a specified range.
func (l *LedgerOperations) getRange(ctx context.Context, op *proto_connector.LedgerOperation) (*proto_connector.Response, error) {

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

	odc := proto_orderer.NewOrdererClient(ordererConn)
	txs, err := odc.GetRange(ctx, &proto_orderer.GetRangeParams{
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

	return &proto_connector.Response{
		ID:     op.GetID(),
		Status: 200,
		Body:   body,
	}, nil
}
