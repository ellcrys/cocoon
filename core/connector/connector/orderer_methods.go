// Package connector methods that call the orderer. This is part of the connector struct
// but is separated to keep the package simple and easy to understand
package connector

import (
	"time"

	"github.com/ellcrys/util"
	"github.com/ncodes/cocoon/core/common"
	"github.com/ncodes/cocoon/core/orderer/proto_orderer"
	"github.com/ncodes/cocoon/core/types"
	context "golang.org/x/net/context"
)

// GetCocoon returns the cocoon being run
func (cn *Connector) GetCocoon(ctx context.Context, cocoonID string) (*types.Cocoon, error) {

	ordererConn, err := cn.ordererDiscovery.GetGRPConn()
	if err != nil {
		return nil, err
	}
	defer ordererConn.Close()

	odc := proto_orderer.NewOrdererClient(ordererConn)
	tx, err := odc.Get(ctx, &proto_orderer.GetParams{
		CocoonID: types.SystemCocoonID,
		Key:      types.MakeCocoonKey(cocoonID),
		Ledger:   types.GetSystemPublicLedgerName(),
	})
	if err != nil {
		if common.CompareErr(err, types.ErrTxNotFound) == 0 {
			return nil, types.ErrCocoonNotFound
		}
		return nil, err
	}

	var cocoon types.Cocoon
	util.FromJSON([]byte(tx.Value), &cocoon)
	return &cocoon, nil
}

// PutCocoon adds a new cocoon. If another cocoon with a matching key
// exists, it is effectively shadowed
func (cn *Connector) PutCocoon(ctx context.Context, cocoon *types.Cocoon) error {

	ordererConn, err := cn.ordererDiscovery.GetGRPConn()
	if err != nil {
		return err
	}
	defer ordererConn.Close()

	createdAt, _ := time.Parse(time.RFC3339Nano, cocoon.CreatedAt)
	odc := proto_orderer.NewOrdererClient(ordererConn)
	_, err = odc.Put(ctx, &proto_orderer.PutTransactionParams{
		CocoonID:   types.SystemCocoonID,
		LedgerName: types.GetSystemPublicLedgerName(),
		Transactions: []*proto_orderer.Transaction{
			&proto_orderer.Transaction{
				Id:        util.UUID4(),
				Key:       types.MakeCocoonKey(cocoon.ID),
				Value:     string(cocoon.ToJSON()),
				CreatedAt: createdAt.Unix(),
			},
		},
	})

	return err
}
