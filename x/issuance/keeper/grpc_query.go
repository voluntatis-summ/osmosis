package keeper

import (
	"github.com/c-osmosis/osmosis/x/issuance/types"
)

var _ types.QueryServer = Keeper{}
