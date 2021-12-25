package keeper

import (
	"context"
	"time"

	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/osmosis-labs/osmosis/x/osmolbp"
	"github.com/osmosis-labs/osmosis/x/osmolbp/proto"
)

var _ proto.MsgServer = Keeper{}

func (k Keeper) CreateLBP(goCtx context.Context, msg *proto.MsgCreateLBP) (*proto.MsgCreateLBPResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	store := ctx.KVStore(k.storeKey)
	id, err := k.createLBP(msg, ctx.BlockTime(), store)
	if err != nil {
		return nil, err
	}
	err = ctx.EventManager().EmitTypedEvent(&proto.EventCreateLBP{
		Id:       id,
		Creator:  msg.Creator,
		TokenIn:  msg.TokenIn,
		TokenOut: msg.TokenOut,
	})
	return &proto.MsgCreateLBPResponse{PoolId: id}, err
}

func (k Keeper) createLBP(msg *proto.MsgCreateLBP, now time.Time, store storetypes.KVStore) (uint64, error) {
	if err := msg.Validate(now); err != nil {
		return 0, err
	}
	id, idBz := k.nextPoolID(store)

	lbpAddr := osmolbp.NewLbpAddress(id)

	p := proto.LBP{
		Id:             id,
		Address:        lbpAddr.String(),
		TokenOut:       msg.TokenOut,
		TokenIn:        msg.TokenIn,
		StartTime:      msg.StartTime,
		EndTime:        msg.StartTime.Add(msg.Duration),
		Rate:           msg.InitialDeposit.Amount.Quo(sdk.NewInt(int64(msg.Duration / proto.ROUND))),
		AccumulatorOut: sdk.ZeroInt(),
		Round:          0,
		Staked:         sdk.ZeroInt(),
	}
	k.savePool(store, idBz, &p)
	return id, nil

}

func (k Keeper) Deposit(goCtx context.Context, msg *proto.MsgDeposit) (*proto.EmptyResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	store := ctx.KVStore(k.storeKey)
	if err := k.deposit(ctx, msg, store); err != nil {
		return nil, err
	}
	err := ctx.EventManager().EmitTypedEvent(&proto.EventDeposit{
		Sender: msg.Sender,
		PoolId: msg.PoolId,
		Amount: msg.Amount.String(),
	})
	return &proto.EmptyResponse{}, err
}

func (k Keeper) deposit(ctx sdk.Context, msg *proto.MsgDeposit, store storetypes.KVStore) error {
	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return err
	}
	lbp, lbpIdBz, err := k.getLBP(store, msg.PoolId)
	if err != nil {
		return err
	}

	if msg.Amount.Denom != lbp.TokenIn {
		return errors.Wrap(errors.ErrInvalidCoins, "deposit denom must be the same as token in denom")
	}

	err = k.bank.SendCoins(ctx, sender, lbp.GetAddress(), sdk.NewCoins(msg.Amount))
	if err != nil {
		return errors.Wrap(err, "user doesn't have enough tokens to stake")
	}

	v, found, err := k.getUserVault(store, lbpIdBz, sender)
	if err != nil {
		return err
	}
	if !found {
		v.Accumulator = lbp.AccumulatorOut
		v.Staked = sdk.ZeroInt()
	}

	stakeInPool(&lbp, &v, msg.Amount.Amount, ctx.BlockTime())

	k.savePool(store, lbpIdBz, &lbp)
	k.saveUserVault(store, lbpIdBz, sender, &v)
	return nil
}

func (k Keeper) Withdraw(goCtx context.Context, msg *proto.MsgWithdraw) (*proto.EmptyResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	store := ctx.KVStore(k.storeKey)
	if err := k.withdraw(ctx, msg, store); err != nil {
		return nil, err
	}
	err := ctx.EventManager().EmitTypedEvent(&proto.EventWithdraw{
		Sender: msg.Sender,
		PoolId: msg.PoolId,
		// TODO: Purchased: ,
	})
	return &proto.EmptyResponse{}, err
}

func (k Keeper) withdraw(ctx sdk.Context, msg *proto.MsgWithdraw, store storetypes.KVStore) error {
	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return err
	}
	p, poolIdBz, err := k.getLBP(store, msg.PoolId)
	if err != nil {
		return err
	}
	v, found, err := k.getUserVault(store, poolIdBz, sender)
	if err != nil {
		return err
	}
	if !found {
		return errors.Wrap(errors.ErrKeyNotFound, "user doesn't have a stake")
	}

	// TODO: check if v.Staked makes sense, maybe we should first ping and evaulate
	if err = unstakeFromPool(&p, &v, v.Staked, ctx.BlockTime()); err != nil {
		return err
	}

	k.savePool(store, poolIdBz, &p)

	return nil
}