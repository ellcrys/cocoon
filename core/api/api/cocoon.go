package api

import (
	"fmt"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/ellcrys/util"
	"github.com/ncodes/cocoon/core/api/api/proto"
	"github.com/ncodes/cocoon/core/common"
	orderer_proto "github.com/ncodes/cocoon/core/orderer/proto"
	"github.com/ncodes/cocoon/core/types"
	"github.com/ncodes/cstructs"
	context "golang.org/x/net/context"
)

var (
	// CocoonStatusStarting indicates a starting service
	CocoonStatusStarting = "starting"

	// CocoonStatusRunning indicates a running service
	CocoonStatusRunning = "running"

	// CocoonStatusStopped indicates a stopped service
	CocoonStatusStopped = "stopped"
)

// makeCocoonKey constructs a cocoon key
func (api *API) makeCocoonKey(id string) string {
	return fmt.Sprintf("cocoon.%s", id)
}

// CreateCocoon creates a cocoon
func (api *API) CreateCocoon(ctx context.Context, req *proto.CreateCocoonRequest) (*proto.Response, error) {

	var err error
	var claims jwt.MapClaims

	if claims, err = api.checkCtxAccessToken(ctx); err != nil {
		return nil, types.ErrInvalidOrExpiredToken
	}

	var cocoon types.Cocoon
	cstructs.Copy(req, &cocoon)
	allowDup := req.OptionAllowDuplicate
	timeCreated, _ := time.Parse(time.RFC3339Nano, req.CreatedAt)
	req = nil

	if err := ValidateCocoon(&cocoon); err != nil {
		return nil, err
	}

	cocoon.IdentityID = claims["identity"].(string)
	cocoon.Status = CocoonStatusStarting

	// add cocoon owner identity if not included
	if !util.InStringSlice(cocoon.Signatories, cocoon.IdentityID) {
		cocoon.Signatories = append(cocoon.Signatories, cocoon.IdentityID)
	}

	if !allowDup {
		_, err = api.GetCocoon(ctx, &proto.GetCocoonRequest{
			ID: cocoon.ID,
		})

		if err != nil && err != types.ErrCocoonNotFound {
			return nil, err
		} else if err != types.ErrCocoonNotFound {
			return nil, fmt.Errorf("cocoon with matching ID already exists")
		}
	}

	// if a link cocoon id is provided, check if the linked cocoon exists
	if len(cocoon.Link) > 0 {
		_, err = api.GetCocoon(ctx, &proto.GetCocoonRequest{
			ID: cocoon.Link,
		})
		if err != nil && err != types.ErrCocoonNotFound {
			return nil, err
		} else if err == types.ErrCocoonNotFound {
			return nil, fmt.Errorf("cannot link to a non-existing cocoon")
		}
	}

	ordererConn, err := api.ordererDiscovery.GetGRPConn()
	if err != nil {
		return nil, err
	}
	defer ordererConn.Close()

	value := cocoon.ToJSON()
	odc := orderer_proto.NewOrdererClient(ordererConn)
	_, err = odc.Put(ctx, &orderer_proto.PutTransactionParams{
		CocoonID:   "",
		LedgerName: types.GetGlobalLedgerName(),
		Transactions: []*orderer_proto.Transaction{
			&orderer_proto.Transaction{
				Id:        util.UUID4(),
				Key:       api.makeCocoonKey(cocoon.ID),
				Value:     string(value),
				CreatedAt: timeCreated.Unix(),
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

// GetCocoon fetches a cocoon
func (api *API) GetCocoon(ctx context.Context, req *proto.GetCocoonRequest) (*proto.Response, error) {

	ordererConn, err := api.ordererDiscovery.GetGRPConn()
	if err != nil {
		return nil, err
	}
	defer ordererConn.Close()

	odc := orderer_proto.NewOrdererClient(ordererConn)
	tx, err := odc.Get(ctx, &orderer_proto.GetParams{
		CocoonID: "",
		Key:      api.makeCocoonKey(req.GetID()),
		Ledger:   types.GetGlobalLedgerName(),
	})

	if err != nil && common.CompareErr(err, types.ErrTxNotFound) != 0 {
		return nil, err
	} else if err != nil && common.CompareErr(err, types.ErrTxNotFound) == 0 {
		return nil, types.ErrCocoonNotFound
	}

	return &proto.Response{
		Status: 200,
		Body:   []byte(tx.GetValue()),
	}, nil
}

// GetCocoonStatus fetches the cocoon status.It queries the scheduler
// to find out the current service status for the cocoon.
func (api *API) GetCocoonStatus(cocoonID string) (string, error) {

	s, err := api.scheduler.GetServiceDiscoverer().GetByID("cocoon", map[string]string{
		"tag": cocoonID,
	})
	if err != nil {
		apiLog.Errorf("failed to query cocoon service status: %s", err.Error())
		return "", fmt.Errorf("failed to query cocoon service status")
	}

	if len(s) == 0 {
		return CocoonStatusStopped, nil
	}

	return CocoonStatusRunning, nil
}
