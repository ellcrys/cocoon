package api

import (
	"fmt"
	"time"

	"github.com/ellcrys/util"
	"github.com/ncodes/cocoon/core/api/api/proto"
	"github.com/ncodes/cocoon/core/common"
	"github.com/ncodes/cocoon/core/orderer"
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
	req = nil

	if err := ValidateRelease(&release); err != nil {
		return nil, err
	}

	ordererConn, err := orderer.DialOrderer(api.orderersAddr)
	if err != nil {
		return nil, err
	}
	defer ordererConn.Close()

	odc := orderer_proto.NewOrdererClient(ordererConn)

	// check if release with matching ID already exists
	ctx, _ = context.WithTimeout(ctx, 2*time.Minute)
	_, err = odc.Get(ctx, &orderer_proto.GetParams{
		CocoonID: "",
		Key:          api.makeReleaseKey(release.ID),
		Ledger:       types.GetGlobalLedgerName(),
	})

	if err != nil && common.CompareErr(err, types.ErrTxNotFound) != 0 {
		return nil, err
	} else if err == nil {
		return nil, fmt.Errorf("a release with matching id already exists")
	}

	value, _ := util.ToJSON(req)
	_, err = odc.Put(ctx, &orderer_proto.PutTransactionParams{
		CocoonID: "",
		LedgerName:   types.GetGlobalLedgerName(),
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
		ID:     release.ID,
		Status: 200,
		Body:   value,
	}, nil
}
