package connector

import (
	"github.com/ncodes/cocoon/core/types"
	"github.com/pkg/errors"

	"golang.org/x/net/context"
)

// persistNetUsage persists the cocoon incoming and outgoing network usage values.
// It returns the current inbound and outbound usage after or error it failed to
// persist theses values
func (cn *Connector) persistNetUsage(ctx context.Context) (uint64, uint64, error) {

	cocoon, err := cn.Platform.GetCocoon(ctx, cn.spec.ID)
	if err != nil {
		return 0, 0, errors.Wrapf(err, "failed to persist network usage: %s", err)
	}

	if cocoon.ResourceUsage == nil {
		cocoon.ResourceUsage = &types.ResourceUsage{}
	}

	// if the latest inbound network is less than the cocoon's inbound network usage,
	// add the latest value to the existing inbound network value. Otherwise, overwrite the
	// cocoon's inbound network usage with the latest
	if cn.resourceUsage.NetRx < cocoon.ResourceUsage.NetIn {
		cocoon.ResourceUsage.NetIn = cocoon.ResourceUsage.NetIn + cn.resourceUsage.NetRx
	} else {
		cocoon.ResourceUsage.NetIn = cn.resourceUsage.NetRx
	}

	// if the latest outbound network is less than the cocoon's outbound network usage,
	// add the latest value to the existing outbound network value. Otherwise, overwrite the
	// cocoon's outbound network usage with the latest
	if cn.resourceUsage.NetTx < cocoon.ResourceUsage.NetOut {
		cocoon.ResourceUsage.NetOut = cocoon.ResourceUsage.NetOut + cn.resourceUsage.NetTx
	} else {
		cocoon.ResourceUsage.NetOut = cn.resourceUsage.NetTx
	}

	if err = cn.Platform.PutCocoon(ctx, cocoon); err != nil {
		return 0, 0, errors.Wrapf(err, "failed update cocoon's network usage: %s", err)
	}

	return cocoon.ResourceUsage.NetIn, cocoon.ResourceUsage.NetOut, nil
}
