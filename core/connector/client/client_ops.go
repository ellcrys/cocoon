package client

import (
	"fmt"
	"strconv"

	"github.com/ellcrys/util"
	"github.com/ncodes/cocoon/core/orderer"
	order_proto "github.com/ncodes/cocoon/core/orderer/proto"
	"github.com/ncodes/cocoon/core/stubs/golang/proto"
	"github.com/ncodes/cocoon/core/types"
	context "golang.org/x/net/context"
)

// createLedger sends a request to the order to create
// a new ledger.
func (c *Client) createLedger(tx *proto.Tx) error {

	ordererConn, err := orderer.DialOrderer(c.orderersAddr)
	if err != nil {
		return err
	}
	defer ordererConn.Close()

	odc := order_proto.NewOrdererClient(ordererConn)
	result, err := odc.CreateLedger(context.Background(), &order_proto.CreateLedgerParams{
		CocoonCodeId: c.cocoonID,
		Name:         tx.GetParams()[0],
		Chained:      tx.GetParams()[1] == "true",
		Public:       tx.GetParams()[2] == "true",
	})
	if err != nil {
		return err
	}

	body, _ := util.ToJSON(result)

	c.stream.Send(&proto.Tx{
		Response: true,
		Status:   200,
		Id:       tx.GetId(),
		Body:     body,
	})

	return nil
}

// getLedger fetches a ledger by its name and cocoon code id
func (c *Client) getLedger(tx *proto.Tx) error {

	ordererConn, err := orderer.DialOrderer(c.orderersAddr)
	if err != nil {
		return err
	}
	defer ordererConn.Close()

	// if name is the global ledger, then a cocoon code id is not required.
	name := tx.GetParams()[0]
	cocoonCodeID := c.cocoonID
	if name == types.GetGlobalLedgerName() {
		cocoonCodeID = ""
	}

	odc := order_proto.NewOrdererClient(ordererConn)
	result, err := odc.GetLedger(context.Background(), &order_proto.GetLedgerParams{
		Name:         name,
		CocoonCodeId: cocoonCodeID,
	})

	if err != nil {
		return err
	}

	body, _ := util.ToJSON(result)

	c.stream.Send(&proto.Tx{
		Response: true,
		Status:   200,
		Id:       tx.GetId(),
		Body:     body,
	})

	return nil
}

// put adds a new transaction to a ledger
func (c *Client) put(tx *proto.Tx) error {

	var txs []*order_proto.Transaction
	ordererConn, err := orderer.DialOrderer(c.orderersAddr)
	if err != nil {
		return err
	}
	defer ordererConn.Close()

	err = util.FromJSON(tx.GetBody(), &txs)
	if err != nil {
		return fmt.Errorf("failed to coerce transaction from bytes to order_proto.Transaction")
	}

	odc := order_proto.NewOrdererClient(ordererConn)
	result, err := odc.Put(context.Background(), &order_proto.PutTransactionParams{
		CocoonCodeId: c.cocoonID,
		LedgerName:   tx.GetParams()[0],
		Transactions: txs,
	})

	if err != nil {
		return err
	}

	body, _ := util.ToJSON(result.Block)

	c.stream.Send(&proto.Tx{
		Response: true,
		Status:   200,
		Id:       tx.GetId(),
		Body:     body,
	})

	return nil
}

// get gets a transaction by its key.
// If byID is set, it will find the transaction by id specified in tx.Id.
func (c *Client) get(tx *proto.Tx, byID bool) error {

	var result *order_proto.Transaction
	var err error

	ordererConn, err := orderer.DialOrderer(c.orderersAddr)
	if err != nil {
		return err
	}
	defer ordererConn.Close()

	odc := order_proto.NewOrdererClient(ordererConn)

	if !byID {
		result, err = odc.Get(context.Background(), &order_proto.GetParams{
			CocoonCodeId: c.cocoonID,
			Ledger:       tx.GetParams()[0],
			Key:          tx.GetParams()[1],
		})
	} else {
		result, err = odc.GetByID(context.Background(), &order_proto.GetParams{
			CocoonCodeId: c.cocoonID,
			Ledger:       tx.GetParams()[0],
			Id:           tx.GetParams()[1],
		})
	}

	if err != nil {
		return err
	}

	body, _ := util.ToJSON(result)

	c.stream.Send(&proto.Tx{
		Response: true,
		Status:   200,
		Id:       tx.GetId(),
		Body:     body,
	})

	return nil
}

// getBlock gets a block by its ledger name and id
func (c *Client) getBlock(tx *proto.Tx) error {

	ordererConn, err := orderer.DialOrderer(c.orderersAddr)
	if err != nil {
		return err
	}
	defer ordererConn.Close()

	odc := order_proto.NewOrdererClient(ordererConn)
	block, err := odc.GetBlockByID(context.Background(), &order_proto.GetBlockParams{
		CocoonCodeId: c.cocoonID,
		Ledger:       tx.GetParams()[0],
		Id:           tx.GetParams()[1],
	})

	if err != nil {
		return err
	}

	body, _ := util.ToJSON(block)

	c.stream.Send(&proto.Tx{
		Response: true,
		Status:   200,
		Id:       tx.GetId(),
		Body:     body,
	})

	return nil
}

// getRange fetches transactions with keys between a specified range.
func (c *Client) getRange(tx *proto.Tx) error {

	ordererConn, err := orderer.DialOrderer(c.orderersAddr)
	if err != nil {
		return err
	}
	defer ordererConn.Close()

	limit, _ := strconv.Atoi(tx.GetParams()[4])
	offset, _ := strconv.Atoi(tx.GetParams()[5])

	odc := order_proto.NewOrdererClient(ordererConn)
	txs, err := odc.GetRange(context.Background(), &order_proto.GetRangeParams{
		CocoonCodeId: c.cocoonID,
		Ledger:       tx.GetParams()[0],
		StartKey:     tx.GetParams()[1],
		EndKey:       tx.GetParams()[2],
		Inclusive:    tx.GetParams()[3] == "true",
		Limit:        int32(limit),
		Offset:       int32(offset),
	})

	if err != nil {
		return err
	}

	body, _ := util.ToJSON(txs.Transactions)

	c.stream.Send(&proto.Tx{
		Response: true,
		Status:   200,
		Id:       tx.GetId(),
		Body:     body,
	})

	return nil
}
