package client

import (
	"github.com/ellcrys/util"
	"github.com/ncodes/cocoon/core/ledgerchain/types"
	"github.com/ncodes/cocoon/core/orderer"
	order_proto "github.com/ncodes/cocoon/core/orderer/proto"
	"github.com/ncodes/cocoon/core/stubs/golang/proto"
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
		Public:       tx.GetParams()[1] == "true",
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

	ordererConn, err := orderer.DialOrderer(c.orderersAddr)
	if err != nil {
		return err
	}
	defer ordererConn.Close()

	odc := order_proto.NewOrdererClient(ordererConn)
	result, err := odc.Put(context.Background(), &order_proto.PutTransactionParams{
		CocoonCodeId: c.cocoonID,
		LedgerName:   tx.GetParams()[0],
		Id:           tx.GetParams()[1],
		Key:          tx.GetParams()[2],
		Value:        []byte(tx.GetParams()[3]),
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

// get gets a transaction by its key
func (c *Client) get(tx *proto.Tx) error {

	ordererConn, err := orderer.DialOrderer(c.orderersAddr)
	if err != nil {
		return err
	}
	defer ordererConn.Close()

	odc := order_proto.NewOrdererClient(ordererConn)
	result, err := odc.Get(context.Background(), &order_proto.GetParams{
		CocoonCodeId: c.cocoonID,
		Ledger:       tx.GetParams()[0],
		Key:          tx.GetParams()[1],
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
