package api

import (
	"fmt"
	"time"

	"github.com/ellcrys/util"
	"github.com/ncodes/cocoon/core/api/api/proto"
	"github.com/ncodes/cocoon/core/common"
	orderer_proto "github.com/ncodes/cocoon/core/orderer/proto"
	"github.com/ncodes/cocoon/core/types"
	"github.com/ncodes/cstructs"
	context "golang.org/x/net/context"
)

// makeIdentityKey constructs an identity key
func (api *API) makeReleaseKey(id string) string {
	return fmt.Sprintf("release.%s", id)
}

// CreateRelease creates a release
func (api *API) CreateRelease(ctx context.Context, req *proto.CreateReleaseRequest) (*proto.Response, error) {

	var release types.Release
	cstructs.Copy(req, &release)
	allowDup := req.OptionAllowDuplicate
	req = nil

	if err := ValidateRelease(&release); err != nil {
		return nil, err
	}

	ordererConn, err := api.ordererDiscovery.GetGRPConn()
	if err != nil {
		return nil, err
	}
	defer ordererConn.Close()

	// check if release with matching ID already exists
	odc := orderer_proto.NewOrdererClient(ordererConn)

	if !allowDup {
		_, err = odc.Get(ctx, &orderer_proto.GetParams{
			CocoonID: "",
			Key:      api.makeReleaseKey(release.ID),
			Ledger:   types.GetGlobalLedgerName(),
		})

		if err != nil && common.CompareErr(err, types.ErrTxNotFound) != 0 {
			return nil, err
		} else if err == nil {
			return nil, fmt.Errorf("a release with matching id already exists")
		}
	}

	value, _ := util.ToJSON(release)
	_, err = odc.Put(ctx, &orderer_proto.PutTransactionParams{
		CocoonID:   "",
		LedgerName: types.GetGlobalLedgerName(),
		Transactions: []*orderer_proto.Transaction{
			&orderer_proto.Transaction{
				Id:        util.UUID4(),
				Key:       api.makeReleaseKey(release.ID),
				Value:     string(value),
				CreatedAt: time.Now().Unix(),
			},
		},
	})
	if err != nil {
		return nil, err
	}

	return &proto.Response{
		Status: 200,
		Body:   value,
	}, nil
}

// GetRelease returns a release
func (api *API) GetRelease(ctx context.Context, req *proto.GetReleaseRequest) (*proto.Response, error) {

	ordererConn, err := api.ordererDiscovery.GetGRPConn()
	if err != nil {
		return nil, err
	}
	defer ordererConn.Close()

	odc := orderer_proto.NewOrdererClient(ordererConn)
	ctx, _ = context.WithTimeout(ctx, 2*time.Minute)
	tx, err := odc.Get(ctx, &orderer_proto.GetParams{
		CocoonID: "",
		Key:      api.makeReleaseKey(req.ID),
		Ledger:   types.GetGlobalLedgerName(),
	})
	if err != nil {
		return nil, err
	}

	return &proto.Response{
		Body:   []byte(tx.GetValue()),
		Status: 200,
	}, nil
}
