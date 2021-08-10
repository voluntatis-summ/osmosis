package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	appparams "github.com/osmosis-labs/osmosis/app/params"
	"github.com/osmosis-labs/osmosis/x/gamm/v1/types"
)

// Parameter store keys
var (
	KeyPoolCreationFee = []byte("PoolCreationFee")
)

// ParamTable for gamm module.
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&types.Params{})
}

func NewParams(poolCreationFee sdk.Coins) types.Params {
	return types.Params{
		PoolCreationFee: poolCreationFee,
	}
}

// default gamm module parameters
func DefaultParams() types.Params {
	return types.Params{
		PoolCreationFee: sdk.Coins{sdk.NewInt64Coin(appparams.BaseCoinUnit, 1000_000_000)}, // 1000 OSMO
	}
}
