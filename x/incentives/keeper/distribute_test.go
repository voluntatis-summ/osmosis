package keeper_test

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/v7/x/incentives/types"
	lockuptypes "github.com/osmosis-labs/osmosis/v7/x/lockup/types"
)

// TestDistribute tests that when the distribute command is executed on
// a provided gauge,
func (suite *KeeperTestSuite) TestDistribute() {
	twoLockupUser := userLocks{
		lockDurations: []time.Duration{defaultLockDuration, 2 * defaultLockDuration},
		lockAmounts:   []sdk.Coins{defaultLPTokens, defaultLPTokens},
	}
	defaultGauge := perpGaugeDesc{
		lockDenom:    defaultLPDenom,
		lockDuration: defaultLockDuration,
		rewardAmount: sdk.Coins{sdk.NewInt64Coin(defaultRewardDenom, 3000)},
	}
	doubleLengthGauge := perpGaugeDesc{
		lockDenom:    defaultLPDenom,
		lockDuration: 2 * defaultLockDuration,
		rewardAmount: sdk.Coins{sdk.NewInt64Coin(defaultRewardDenom, 3000)},
	}
	oneKRewardCoins := sdk.Coins{sdk.NewInt64Coin(defaultRewardDenom, 1000)}
	twoKRewardCoins := sdk.Coins{sdk.NewInt64Coin(defaultRewardDenom, 2000)}
	fiveKRewardCoins := sdk.Coins{sdk.NewInt64Coin(defaultRewardDenom, 5000)}
	tests := []struct {
		users           []userLocks
		gauges          []perpGaugeDesc
		expectedRewards []sdk.Coins
	}{
		// gauge 1 gives 3k coins. Three locks, all eligible. 1k coins per lock
		// so 1k to oneLockupUser, 2k to twoLockupUser
		{
			users:           []userLocks{oneLockupUser, twoLockupUser},
			gauges:          []perpGaugeDesc{defaultGauge},
			expectedRewards: []sdk.Coins{oneKRewardCoins, twoKRewardCoins},
		},
		// gauge 1 gives 3k coins. Three locks, all eligible.
		// gauge 2 gives 3k coins to one lock, in twoLockupUser
		// so 1k to oneLockupUser, 5k to twoLockupUser
		{
			users:           []userLocks{oneLockupUser, twoLockupUser},
			gauges:          []perpGaugeDesc{defaultGauge, doubleLengthGauge},
			expectedRewards: []sdk.Coins{oneKRewardCoins, fiveKRewardCoins},
		},
	}
	for tcIndex, tc := range tests {
		suite.SetupTest()
		gauges := suite.SetupGauges(tc.gauges)
		addrs := suite.SetupUserLocks(tc.users)
		_, err := suite.app.IncentivesKeeper.Distribute(suite.ctx, gauges)
		suite.Require().NoError(err)
		// Check expected rewards
		for i, addr := range addrs {
			bal := suite.app.BankKeeper.GetAllBalances(suite.ctx, addr)
			suite.Require().Equal(tc.expectedRewards[i].String(), bal.String(), "tcnum %d, person %d", tcIndex, i)
		}
	}

	// TODO: test distribution for synthetic lockup as well
}

type gauge struct {
	isPerpetual bool
	isLock      bool
	coinsToAdd  sdk.Coins
}

func (suite *KeeperTestSuite) TestGetModuleToDistributeCoins() {
	testcases := []struct {
		desc        string
		gauges      []gauge
		isPerpetual bool
	}{
		{
			desc:        "Test lock perpetual gauge distribution",
			isPerpetual: false,
			gauges: []gauge{
				{
					isLock:      true,
					isPerpetual: false,
					coinsToAdd:  sdk.Coins{sdk.NewInt64Coin("stake", 200)},
				},
			},
		},
		{
			desc:        "Test lock perpetual gauge distribution with second gauge",
			isPerpetual: false,
			gauges: []gauge{
				{
					isLock:      true,
					isPerpetual: false,
					coinsToAdd:  sdk.Coins{sdk.NewInt64Coin("stake", 200)},
				},
				{
					isLock:      false,
					isPerpetual: false,
					coinsToAdd:  sdk.Coins{sdk.NewInt64Coin("stake", 1000)},
				},
			},
		},
		{
			desc:        "Test no lock perpetual gauge distribution",
			isPerpetual: true,
			gauges: []gauge{
				{
					isLock:      false,
					isPerpetual: true,
					coinsToAdd:  nil,
				},
			},
		},
		{
			desc:        "Test no lock & non perpetual gauge distribution",
			isPerpetual: false,
			gauges: []gauge{
				{
					isLock:      false,
					isPerpetual: false,
					coinsToAdd:  nil,
				},
			},
		},
		{
			desc:        "Test no lock perpetual gauge distribution",
			isPerpetual: true,
			gauges: []gauge{
				{
					isLock:      false,
					isPerpetual: true,
					coinsToAdd:  nil,
				},
			},
		},
		{
			desc:        "Test no lock & non perpetual gauge distribution",
			isPerpetual: false,
			gauges: []gauge{
				{
					isLock:      false,
					isPerpetual: false,
					coinsToAdd:  nil,
				},
			},
		},
	}

	for _, tc := range testcases {
		suite.Run(tc.desc, func() {
			// test for module get gauges
			suite.SetupTest()

			// initial check
			suite.validateIncentivesModuleInitialization()

			// coinsFromGauges running total of coins from gauges
			// and addition to gauges
			var coinsFromGauges sdk.Coins

			// gaugeDetailsMap used to track gauge cretion details
			var gaugeDetailsMap = gaugeDetailsMap{}

			// loop thorugh gauges checking creation and addition
			gaugeDetailsMap, coinsFromGauges = suite.validateTestcaseGauges(tc.gauges, tc.isPerpetual, gaugeDetailsMap, coinsFromGauges)

			// test upcoming gauges before distribution
			suite.validateUpcomingGaugesBeforeDistribution(tc.gauges)

			// start distribution
			suite.validateDistribution(tc.gauges, gaugeDetailsMap, tc.isPerpetual, coinsFromGauges)
		})
	}
}
