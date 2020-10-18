package keeper

import (
	"github.com/c-osmosis/osmosis/x/gamm/keeper/pool"
	"github.com/c-osmosis/osmosis/x/gamm/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
)

var _ Keeper = (*keeper)(nil)

type Keeper interface {
	pool.Service
}

type keeper struct {
	pool.Service

	cdc           codec.BinaryMarshaler
	storeKey      sdk.StoreKey
	accountKeeper types.AccountKeeper
	bankKeeper    bankkeeper.Keeper
}

func NewKeeper(cdc codec.BinaryMarshaler, storeKey sdk.StoreKey, accountKeeper types.AccountKeeper, bankKeeper bankkeeper.Keeper) Keeper {
	return keeper{
		Service: pool.NewService(cdc, storeKey, accountKeeper, bankKeeper),

		cdc:           cdc,
		storeKey:      storeKey,
		accountKeeper: accountKeeper,
		bankKeeper:    bankKeeper,
	}
}