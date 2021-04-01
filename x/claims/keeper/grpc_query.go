package keeper

import (
	"github.com/c-osmosis/osmosis/x/claims/types"
)

var _ types.QueryServer = Keeper{}
