# Superfluid staking

Superfluid staking is to use the OSMO tokens put on liquidity pools to be used both for swap & staking.

## Potential problems

### Is it secure to use it both for staking and swap?

In current design, liquidity is not used enough and therefore, using it both for swap and staking won't be a problem.
As we update the pools to use liquidity more efficiently, almost all of those liquidity could be used for swap.
In that case, the amount of OSMO in the pool does not participate in increasing the security of the chain.

### How to determine the amount of OSMO delegation from the pool to validators?

The amount of OSMO in the pool is being changed in real-time based on swap events.
In staking, the incentives amount record is managed per withdrawal and slash events.
If we invent OSMO tokens in the pool as well, how often should we update the balance of OSMO into the "delegation"?

### Should we actually decide which validator to delegate from the pool?

On slash event, OSMO will be burnt from the pool directly? It won't cause a problem in the gamm for coin amount inconsistency?
Like the k = x * y would be changed

### The reward from staking, it should be withdrawn to the user's wallet on each epoch with yield farming incentives or withdraw it by doing claim?

### Do we really need to select validator for this staking?

What if we think that they are constantly supporting the security of chain and provide specific percentage of fee pool to that rather than dividing by OSMO put the liquidity and the OSMO delegated?

### When the validator is slashed should we burn LP token or burn OSMO inside the pool?

I think burning LP token would be nicer and pretty easier.
LP token's value could be calculated based on how much OSMO is in it.

- How do we prevent a hacker buy/sell just before epoch end and do opposite operation just after epoch?

### To start superfluid staking, users should lock LP tokens for staking withdrawal time? 

### How to prevent a validator to slash itself to hack pools?

### Superfluid staking will be applied for only secure pairs like ATOM/OSMO or AKT/OSMO

### Base thoughts of how OSMO-X LP token pair secures Osmosis chain

OSMO-X LPs secures OSMO token price by providing worth of X tokens to OSMO pair.
It prevents OSMO token price from going down easily.

The TVL on staking is the security of the chain.
The TVL on liquidity is the security of OSMO token price.

Therefore, the incentives could be provided in the same way with stakers.

What stakers get = OSMO tokens from inflation + OSMO tokens on transaction fee
What LPs get = OSMO yield farming from inflation + trading fees - impermanent loss

What LPs don't get here is transaction fees.
Should we actually provide inflation allocated for stakers to LPs? Won't it let the OSMO stakers leave to become LPs?

Or when we implement superfluid staking, should we just remove allocation percentage for pool-incentives?
