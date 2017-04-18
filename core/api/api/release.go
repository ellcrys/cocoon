package api

import (
	"fmt"
	"time"

	"github.com/ellcrys/util"
	"github.com/ncodes/cocoon/core/api/api/proto_api"
	"github.com/ncodes/cocoon/core/common"
	"github.com/ncodes/cocoon/core/orderer/proto_orderer"
	"github.com/ncodes/cocoon/core/types"
	"github.com/ncodes/cstructs"
	context "golang.org/x/net/context"
)

// putRelease adds a new release. If another release with a matching key
// exists, it is effectively shadowed
func (api *API) putRelease(ctx context.Context, release *types.Release) error {

	ordererConn, err := api.ordererDiscovery.GetGRPConn()
	if err != nil {
		return err
	}
	defer ordererConn.Close()

	createdAt, _ := time.Parse(time.RFC3339Nano, release.CreatedAt)
	odc := proto_orderer.NewOrdererClient(ordererConn)
	_, err = odc.Put(ctx, &proto_orderer.PutTransactionParams{
		CocoonID:   types.SystemCocoonID,
		LedgerName: types.GetSystemPublicLedgerName(),
		Transactions: []*proto_orderer.Transaction{
			&proto_orderer.Transaction{
				Id:        util.UUID4(),
				Key:       types.MakeReleaseKey(release.ID),
				Value:     string(release.ToJSON()),
				CreatedAt: createdAt.Unix(),
			},
		},
	})

	return err
}

// CreateRelease creates a release
func (api *API) CreateRelease(ctx context.Context, req *proto_api.CreateReleaseRequest) (*proto_api.Response, error) {

	var release types.Release
	cstructs.Copy(req, &release)
	release.CreatedAt = time.Now().UTC().Format(time.RFC3339Nano)

	// recreate ACLMap from byte value in request
	if len(req.ACL) > 0 {
		var aclMap map[string]interface{}
		if err := util.FromJSON(req.ACL, &aclMap); err != nil {
			return nil, fmt.Errorf("acl: malformed json")
		}
		release.ACL = types.NewACLMap(aclMap)
	}

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

	return &proto_api.Response{
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

	odc := proto_orderer.NewOrdererClient(ordererConn)
	tx, err := odc.Get(ctx, &proto_orderer.GetParams{
		CocoonID: types.SystemCocoonID,
		Key:      types.MakeReleaseKey(id),
		Ledger:   types.GetSystemPublicLedgerName(),
	})
	if err != nil {
		return nil, err
	}

	var release types.Release
	util.FromJSON([]byte(tx.GetValue()), &release)

	return &release, nil
}

// GetRelease returns a release
func (api *API) GetRelease(ctx context.Context, req *proto_api.GetReleaseRequest) (*proto_api.Response, error) {

	release, err := api.getRelease(ctx, req.ID)
	if err != nil {
		return nil, err
	}

	return &proto_api.Response{
		Body:   []byte(release.ToJSON()),
		Status: 200,
	}, nil
}
