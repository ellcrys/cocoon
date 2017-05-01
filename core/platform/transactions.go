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
)

// Transactions represents a collection of
// methods for performing common, orderer-specific operations
type Transactions struct {
	ordererDiscoverer *orderer.Discovery
}

// NewTransactions creates a new transaction object
func NewTransactions() (*Transactions, error) {
	ordererDiscoverer, err := orderer.NewDiscovery()
	if err != nil {
		return nil, err
	}
	go ordererDiscoverer.Discover()
	return &Transactions{
		ordererDiscoverer: ordererDiscoverer,
	}, nil
}

// Stop stops the orderer discoverer service
func (t *Transactions) Stop() {
	t.ordererDiscoverer.Stop()
}

// GetIdentity gets an existing identity and returns an identity object.
// Since an identity password is never saved along side the rest of the other identity
// field on the system public ledger, it is retrieved from the private system ledger where
// it is stored separately.
func (t *Transactions) GetIdentity(ctx context.Context, id string) (*types.Identity, error) {

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

	passwordTx, err := odc.Get(ctx, &proto_orderer.GetParams{
		CocoonID: types.SystemCocoonID,
		Key:      types.MakeIdentityPasswordKey(id),
		Ledger:   types.GetSystemPrivateLedgerName(),
	})
	if err != nil {
		if common.CompareErr(err, types.ErrTxNotFound) == 0 {
			return nil, fmt.Errorf("identity password transaction not found")
		}
		return nil, err
	}

	var identity types.Identity
	util.FromJSON([]byte(tx.GetValue()), &identity)
	identity.Password = passwordTx.GetValue()

	return &identity, nil
}

// PutIdentity adds a new identity. If another identity with a matching key
// exists, it is effectively shadowed. The identity password is not saved to the
// systems public ledger but on the system private ledger.
func (t *Transactions) PutIdentity(ctx context.Context, identity *types.Identity) error {

	var password = identity.Password

	ordererConn, err := t.ordererDiscoverer.GetGRPConn()
	if err != nil {
		return err
	}
	defer ordererConn.Close()

	identity.Password = ""
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

	// save identity password alone in the system private ledger
	_, err = odc.Put(ctx, &proto_orderer.PutTransactionParams{
		CocoonID:   types.SystemCocoonID,
		LedgerName: types.GetSystemPrivateLedgerName(),
		Transactions: []*proto_orderer.Transaction{
			&proto_orderer.Transaction{
				Id:        util.UUID4(),
				Key:       types.MakeIdentityPasswordKey(identity.GetID()),
				Value:     password,
				CreatedAt: createdAt.Unix(),
			},
		},
	})

	return err
}

// PutCocoon adds a new cocoon. If another cocoon with a matching key
// exists, it is effectively shadowed
func (t *Transactions) PutCocoon(ctx context.Context, cocoon *types.Cocoon) error {

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
func (t *Transactions) GetCocoon(ctx context.Context, id string) (*types.Cocoon, error) {

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

// GetCocoonAndLastRelease fetches a cocoon and the last release that was added to it.
// If lastDeployed is set, it is forced to return the last release that was deployed otherwise,
// error is returned. Set includePrivateFields to true to return include private fields of the cocoon and release
func (t *Transactions) GetCocoonAndLastRelease(ctx context.Context, cocoonID string, lastDeployed, includePrivateFields bool) (*types.Cocoon, *types.Release, error) {

	cocoon, err := t.GetCocoon(ctx, cocoonID)
	if err != nil {
		return nil, nil, err
	}

	var releaseID = cocoon.LastDeployedReleaseID
	if lastDeployed && len(cocoon.LastDeployedReleaseID) == 0 {
		return nil, nil, fmt.Errorf("cocoon has no recently deployed release")
	}
	if !lastDeployed && len(cocoon.Releases) > 0 {
		releaseID = cocoon.Releases[len(cocoon.Releases)-1]
	}
	if len(cocoon.Releases) == 0 {
		return nil, nil, fmt.Errorf("cocoon has no release. Wierd")
	}

	release, err := t.GetRelease(ctx, releaseID, includePrivateFields)
	if err != nil {
		return nil, nil, err
	}

	return cocoon, release, nil
}

// PutRelease adds a new release. If another release with a matching key
// exists, it is effectively shadowed
func (t *Transactions) PutRelease(ctx context.Context, release *types.Release) error {

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

	_, err = odc.Put(ctx, &proto_orderer.PutTransactionParams{
		CocoonID:   types.SystemCocoonID,
		LedgerName: types.GetSystemPrivateLedgerName(),
		Transactions: []*proto_orderer.Transaction{
			&proto_orderer.Transaction{
				Id:        util.UUID4(),
				Key:       types.MakeReleaseEnvKey(release.ID),
				Value:     string(privEnv.ToJSON()),
				CreatedAt: createdAt.Unix(),
			},
		},
	})

	return err
}

// GetRelease gets an existing release and returns a release object.
// Passing true to includePrivateFields will fetch other field values stored
// on the private system ledger.
func (t *Transactions) GetRelease(ctx context.Context, id string, includePrivateFields bool) (*types.Release, error) {

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

		// get private environment variables
		privEnvTx, err := odc.Get(ctx, &proto_orderer.GetParams{
			CocoonID: types.SystemCocoonID,
			Key:      types.MakeReleaseEnvKey(id),
			Ledger:   types.GetSystemPrivateLedgerName(),
		})
		if err != nil && common.CompareErr(err, types.ErrTxNotFound) != 0 {
			return nil, err
		}

		// include private environment variables with public variables in the release
		if privEnvTx != nil {
			var privEnvs map[string]string
			util.FromJSON([]byte(privEnvTx.Value), &privEnvs)
			releaseEnvAsMap := release.Env.ToMap()
			mergo.MergeWithOverwrite(&releaseEnvAsMap, privEnvs)
		}
	}

	return &release, nil
}
