package types

type GammSubKeeper interface {
	MinPoolAssets() uint64
	MaxPoolAssets() uint64
}
