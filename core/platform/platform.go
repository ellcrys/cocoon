package platform

import (
	"fmt"
	"time"

	context "golang.org/x/net/context"

	"github.com/ellcrys/util"
	"github.com/imdario/mergo"
	"github.com/ncodes/cocoon/core/common"
	"github.com/ncodes/cocoon/core/orderer/orderer"
	"github.com/ncodes/cocoon/core/orderer/proto_orderer"
	"github.com/ncodes/cocoon/core/types"
	"github.com/pkg/errors"
)

// Platform represents a collection of
// methods for performing common platform business logic
// that involves identity, cocoons and other platform resources
type Platform struct {
	ordererDiscoverer *orderer.Discovery
}

// NewPlatform creates a new transaction object
func NewPlatform() (*Platform, error) {
	ordererDiscoverer, err := orderer.NewDiscovery()
	if err != nil {
		return nil, err
	}
	go ordererDiscoverer.Discover()
	return &Platform{
		ordererDiscoverer: ordererDiscoverer,
	}, nil
}

// GetOrdererDiscoverer returns the orderer discover used
func (t *Platform) GetOrdererDiscoverer() *orderer.Discovery {
	return t.ordererDiscoverer
}

// Stop stops the orderer discoverer service
func (t *Platform) Stop() {
	t.ordererDiscoverer.Stop()
}

// GetIdentity gets an existing identity and returns an identity object.
// Since an identity password is never saved along side the rest of the other identity
// field on the system public ledger, it is retrieved from the private system ledger where
// it is stored separately.
func (t *Platform) GetIdentity(ctx context.Context, id string) (*types.Identity, error) {

	ordererConn, err := t.ordererDiscoverer.GetGRPConn()
	if err != nil {
		return nil, err
	}
	defer ordererConn.Close()

	odc := proto_orderer.NewOrdererClient(ordererConn)
	tx, err := odc.Get(ctx, &proto_orderer.GetParams{
		CocoonID: types.SystemCocoonID,
		Key:      types.MakeIdentityKey(id),
		Ledger:   types.GetSystemPublicLedgerName(),
	})

	if err != nil {
		if common.CompareErr(err, types.ErrTxNotFound) == 0 {
			return nil, types.ErrIdentityNotFound
		}
		return nil, err
	}

	// get private data
	privTx, err := odc.Get(ctx, &proto_orderer.GetParams{
		CocoonID: types.SystemCocoonID,
		Key:      types.MakePrivateIdentityKey(id),
		Ledger:   types.GetSystemPrivateLedgerName(),
	})
	if err != nil {
		if common.CompareErr(err, types.ErrTxNotFound) == 0 {
			return nil, fmt.Errorf("identity's private data not found")
		}
		return nil, err
	}

	var identity types.Identity
	util.FromJSON([]byte(tx.GetValue()), &identity)
	util.FromJSON([]byte(privTx.GetValue()), &identity)

	return &identity, nil
}

// PutIdentity adds a new identity. If another identity with a matching key
// exists, it is effectively shadowed. The identity password is not saved to the
// systems public ledger but on the system private ledger.
func (t *Platform) PutIdentity(ctx context.Context, identity *types.Identity) error {

	ordererConn, err := t.ordererDiscoverer.GetGRPConn()
	if err != nil {
		return err
	}
	defer ordererConn.Close()

	var password = identity.Password
	var clientSession = identity.ClientSessions

	// remove private data
	identity.Password = ""
	identity.ClientSessions = nil

	createdAt, _ := time.Parse(time.RFC3339Nano, identity.CreatedAt)
	odc := proto_orderer.NewOrdererClient(ordererConn)
	_, err = odc.Put(ctx, &proto_orderer.PutTransactionParams{
		CocoonID:   types.SystemCocoonID,
		LedgerName: types.GetSystemPublicLedgerName(),
		Transactions: []*proto_orderer.Transaction{
			&proto_orderer.Transaction{
				Id:        util.UUID4(),
				Key:       types.MakeIdentityKey(identity.GetID()),
				Value:     string(identity.ToJSON()),
				CreatedAt: createdAt.Unix(),
			},
		},
	})

	// create new identity object for private data
	var privateIdentity = &types.Identity{
		Password:       password,
		ClientSessions: clientSession,
	}

	// save identity password alone in the system private ledger
	_, err = odc.Put(ctx, &proto_orderer.PutTransactionParams{
		CocoonID:   types.SystemCocoonID,
		LedgerName: types.GetSystemPrivateLedgerName(),
		Transactions: []*proto_orderer.Transaction{
			&proto_orderer.Transaction{
				Id:        util.UUID4(),
				Key:       types.MakePrivateIdentityKey(identity.GetID()),
				Value:     string(privateIdentity.ToJSON()),
				CreatedAt: createdAt.Unix(),
			},
		},
	})

	return err
}

// PutCocoon adds a new cocoon. If another cocoon with a matching key
// exists, it is effectively shadowed
func (t *Platform) PutCocoon(ctx context.Context, cocoon *types.Cocoon) error {

	ordererConn, err := t.ordererDiscoverer.GetGRPConn()
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

// GetCocoon gets a cocoon with a matching id.
func (t *Platform) GetCocoon(ctx context.Context, id string) (*types.Cocoon, error) {

	ordererConn, err := t.ordererDiscoverer.GetGRPConn()
	if err != nil {
		return nil, err
	}
	defer ordererConn.Close()

	odc := proto_orderer.NewOrdererClient(ordererConn)
	tx, err := odc.Get(ctx, &proto_orderer.GetParams{
		CocoonID: types.SystemCocoonID,
		Key:      types.MakeCocoonKey(id),
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

// GetCocoonAndRelease fetches a cocoon and a release that once. Returns error if either of them are not found.
func (t *Platform) GetCocoonAndRelease(ctx context.Context, cocoonID, releaseID string, includePrivateFields bool) (*types.Cocoon, *types.Release, error) {

	cocoon, err := t.GetCocoon(ctx, cocoonID)
	if err != nil {
		if common.CompareErr(err, types.ErrTxNotFound) == 0 {
			return nil, nil, errors.Wrap(err, "cocoon not found")
		}
		return nil, nil, err
	}

	release, err := t.GetRelease(ctx, releaseID, includePrivateFields)
	if err != nil {
		if common.CompareErr(err, types.ErrTxNotFound) == 0 {
			return nil, nil, errors.Wrap(err, "release not found")
		}
		return nil, nil, err
	}

	return cocoon, release, nil
}

// GetCocoonAndLastActiveRelease fetches a cocoon with the release that was last deployed or last created.
// Set includePrivateFields to true to include private fields of the cocoon and release
func (t *Platform) GetCocoonAndLastActiveRelease(ctx context.Context, cocoonID string, includePrivateFields bool) (*types.Cocoon, *types.Release, error) {

	cocoon, err := t.GetCocoon(ctx, cocoonID)
	if err != nil {
		return nil, nil, err
	}

	var releaseID = cocoon.LastDeployedReleaseID
	if len(cocoon.LastDeployedReleaseID) == 0 {
		if len(cocoon.Releases) == 0 {
			return nil, nil, errors.Wrap(err, "cocoon has no release. Wierd")
		}
		releaseID = cocoon.Releases[len(cocoon.Releases)-1]
	}

	release, err := t.GetRelease(ctx, releaseID, includePrivateFields)
	if err != nil {
		return nil, nil, err
	}

	return cocoon, release, nil
}

// PutRelease adds a new release. If another release with a matching key
// exists, it is effectively shadowed
func (t *Platform) PutRelease(ctx context.Context, release *types.Release) error {

	ordererConn, err := t.ordererDiscoverer.GetGRPConn()
	if err != nil {
		return err
	}
	defer ordererConn.Close()
	pubEnv, privEnv := release.Env.Process(true)
	release.Env = pubEnv

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

	var privRelease = &types.Release{
		Env: privEnv,
	}

	_, err = odc.Put(ctx, &proto_orderer.PutTransactionParams{
		CocoonID:   types.SystemCocoonID,
		LedgerName: types.GetSystemPrivateLedgerName(),
		Transactions: []*proto_orderer.Transaction{
			&proto_orderer.Transaction{
				Id:        util.UUID4(),
				Key:       types.MakePrivateReleaseKey(release.ID),
				Value:     string(privRelease.ToJSON()),
				CreatedAt: createdAt.Unix(),
			},
		},
	})

	return err
}

// GetRelease gets an existing release and returns a release object.
// Passing true to includePrivateFields will fetch other field values stored
// on the private system ledger.
func (t *Platform) GetRelease(ctx context.Context, id string, includePrivateFields bool) (*types.Release, error) {

	ordererConn, err := t.ordererDiscoverer.GetGRPConn()
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

	if includePrivateFields {

		// get private release data
		privTx, err := odc.Get(ctx, &proto_orderer.GetParams{
			CocoonID: types.SystemCocoonID,
			Key:      types.MakePrivateReleaseKey(id),
			Ledger:   types.GetSystemPrivateLedgerName(),
		})
		if err != nil && common.CompareErr(err, types.ErrTxNotFound) != 0 {
			return nil, err
		}

		// include private environment variables with public variables in the release
		if privTx != nil {
			var privateRelease types.Release
			util.FromJSON([]byte(privTx.GetValue()), &privateRelease)
			releaseEnvAsMap := release.Env.ToMap()
			mergo.MergeWithOverwrite(&releaseEnvAsMap, privateRelease.Env.ToMap())
		}
	}

	return &release, nil
}
