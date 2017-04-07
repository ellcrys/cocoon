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

// putRelease adds a new release. If another release with a matching key
// exists, it is effectively shadowed
func (api *API) putRelease(ctx context.Context, release *types.Release) error {

	ordererConn, err := api.ordererDiscovery.GetGRPConn()
	if err != nil {
		return err
	}
	defer ordererConn.Close()

	createdAt, _ := time.Parse(time.RFC3339Nano, release.CreatedAt)
	odc := orderer_proto.NewOrdererClient(ordererConn)
	_, err = odc.Put(ctx, &orderer_proto.PutTransactionParams{
		CocoonID:   "",
		LedgerName: types.GetGlobalLedgerName(),
		Transactions: []*orderer_proto.Transaction{
			&orderer_proto.Transaction{
				Id:        util.UUID4(),
				Key:       api.makeReleaseKey(release.ID),
				Value:     string(release.ToJSON()),
				CreatedAt: createdAt.Unix(),
			},
		},
	})

	return err
}

// CreateRelease creates a release
func (api *API) CreateRelease(ctx context.Context, req *proto.CreateReleaseRequest) (*proto.Response, error) {

	var release types.Release
	cstructs.Copy(req, &release)
	release.CreatedAt = time.Now().UTC().Format(time.RFC3339Nano)
	req = nil

	if err := ValidateRelease(&release); err != nil {
		return nil, err
	}

	_, err := api.getRelease(ctx, release.ID)
	if err != nil && common.CompareErr(err, types.ErrTxNotFound) != 0 {
		return nil, err
	} else if err == nil {
		return nil, fmt.Errorf("a release with matching id already exists")
	}

	err = api.putRelease(ctx, &release)
	if err != nil {
		return nil, err
	}

	return &proto.Response{
		Status: 200,
		Body:   release.ToJSON(),
	}, nil
}

// getRelease gets an existing release and returns a release object
func (api *API) getRelease(ctx context.Context, id string) (*types.Release, error) {

	ordererConn, err := api.ordererDiscovery.GetGRPConn()
	if err != nil {
		return nil, err
	}
	defer ordererConn.Close()

	odc := orderer_proto.NewOrdererClient(ordererConn)
	tx, err := odc.Get(ctx, &orderer_proto.GetParams{
		CocoonID: "",
		Key:      api.makeReleaseKey(id),
		Ledger:   types.GetGlobalLedgerName(),
	})
	if err != nil {
		return nil, err
	}

	var release types.Release
	util.FromJSON([]byte(tx.GetValue()), &release)

	return &release, nil
}

// GetRelease returns a release
func (api *API) GetRelease(ctx context.Context, req *proto.GetReleaseRequest) (*proto.Response, error) {

	release, err := api.getRelease(ctx, req.ID)
	if err != nil {
		return nil, err
	}

	return &proto.Response{
		Body:   []byte(release.ToJSON()),
		Status: 200,
	}, nil
}
