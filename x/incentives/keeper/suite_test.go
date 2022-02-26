package keeper_test

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/v7/x/incentives/types"
	lockuptypes "github.com/osmosis-labs/osmosis/v7/x/lockup/types"
)

var (
	defaultLPDenom      string        = "lptoken"
	defaultLPTokens     sdk.Coins     = sdk.Coins{sdk.NewInt64Coin(defaultLPDenom, 10)}
	defaultLiquidTokens sdk.Coins     = sdk.Coins{sdk.NewInt64Coin("foocoin", 10)}
	defaultLockDuration time.Duration = time.Second
	oneLockupUser       userLocks     = userLocks{
		lockDurations: []time.Duration{time.Second},
		lockAmounts:   []sdk.Coins{defaultLPTokens},
	}
	defaultRewardDenom string = "rewardDenom"
)

// TODO: Switch more code to use userLocks and perpGaugeDesc
type userLocks struct {
	lockDurations []time.Duration
	lockAmounts   []sdk.Coins
}

type perpGaugeDesc struct {
	lockDenom    string
	lockDuration time.Duration
	rewardAmount sdk.Coins
}

type gaugeDetailsMap map[int]gaugeDetails
type gaugeDetails struct {
	gaugeID        uint64
	gaugeCoins     sdk.Coins
	gaugeStartTime time.Time
}

// Leave prefix blank if lazy, it'll be replaced with something random
func (suite *KeeperTestSuite) setupAddr(addrNum int, prefix string, balance sdk.Coins) sdk.AccAddress {
	if prefix == "" {
		prefixBz := make([]byte, 8)
		_, _ = rand.Read(prefixBz)
		prefix = string(prefixBz)
	} else {
		prefix = fmt.Sprintf("%8.8s", prefix)
	}

	addr := sdk.AccAddress([]byte(fmt.Sprintf("addr%s%8d", prefix, addrNum)))
	err := simapp.FundAccount(suite.app.BankKeeper, suite.ctx, addr, balance)
	suite.Require().NoError(err)
	return addr
}

func (suite *KeeperTestSuite) SetupUserLocks(users []userLocks) (accs []sdk.AccAddress) {
	accs = make([]sdk.AccAddress, len(users))
	for i, user := range users {
		suite.Assert().Equal(len(user.lockDurations), len(user.lockAmounts))
		totalLockAmt := user.lockAmounts[0]
		for j := 1; j < len(user.lockAmounts); j++ {
			totalLockAmt = totalLockAmt.Add(user.lockAmounts[j]...)
		}
		accs[i] = suite.setupAddr(i, "", totalLockAmt)
		for j := 0; j < len(user.lockAmounts); j++ {
			_, err := suite.app.LockupKeeper.LockTokens(
				suite.ctx, accs[i], user.lockAmounts[j], user.lockDurations[j])
			suite.Require().NoError(err)
		}
	}
	return
}

func (suite *KeeperTestSuite) SetupGauges(gaugeDescriptors []perpGaugeDesc) []types.Gauge {
	gauges := make([]types.Gauge, len(gaugeDescriptors))
	perpetual := true
	for i, desc := range gaugeDescriptors {
		_, gaugePtr, _, _ := suite.setupNewGaugeWithDuration(perpetual, desc.rewardAmount, desc.lockDuration)
		gauges[i] = *gaugePtr
	}
	return gauges
}

func (suite *KeeperTestSuite) CreateGauge(isPerpetual bool, addr sdk.AccAddress, coins sdk.Coins, distrTo lockuptypes.QueryCondition, startTime time.Time, numEpoch uint64) (uint64, *types.Gauge) {
	err := simapp.FundAccount(suite.app.BankKeeper, suite.ctx, addr, coins)
	suite.Require().NoError(err)
	gaugeID, err := suite.app.IncentivesKeeper.CreateGauge(suite.ctx, isPerpetual, addr, coins, distrTo, startTime, numEpoch)
	suite.Require().NoError(err)
	gauge, err := suite.app.IncentivesKeeper.GetGaugeByID(suite.ctx, gaugeID)
	suite.Require().NoError(err)
	return gaugeID, gauge
}

func (suite *KeeperTestSuite) AddToGauge(coins sdk.Coins, gaugeID uint64) uint64 {
	addr := sdk.AccAddress([]byte("addrx---------------"))
	err := simapp.FundAccount(suite.app.BankKeeper, suite.ctx, addr, coins)
	suite.Require().NoError(err)
	err = suite.app.IncentivesKeeper.AddToGaugeRewards(suite.ctx, addr, coins, gaugeID)
	suite.Require().NoError(err)
	return gaugeID
}

func (suite *KeeperTestSuite) LockTokens(addr sdk.AccAddress, coins sdk.Coins, duration time.Duration) {
	err := simapp.FundAccount(suite.app.BankKeeper, suite.ctx, addr, coins)
	suite.Require().NoError(err)
	_, err = suite.app.LockupKeeper.LockTokens(suite.ctx, addr, coins, duration)
	suite.Require().NoError(err)
}

func (suite *KeeperTestSuite) setupNewGaugeWithDuration(isPerpetual bool, coins sdk.Coins, duration time.Duration) (
	uint64, *types.Gauge, sdk.Coins, time.Time) {
	addr := sdk.AccAddress([]byte("Gauge_Creation_Addr_"))
	startTime2 := time.Now()
	distrTo := lockuptypes.QueryCondition{
		LockQueryType: lockuptypes.ByDuration,
		Denom:         "lptoken",
		Duration:      duration,
	}

	// mints coins so supply exists on chain
	mintCoins := sdk.Coins{sdk.NewInt64Coin(distrTo.Denom, 200)}
	err := simapp.FundAccount(suite.app.BankKeeper, suite.ctx, addr, mintCoins)
	suite.Require().NoError(err)

	numEpochsPaidOver := uint64(2)
	if isPerpetual {
		numEpochsPaidOver = uint64(1)
	}
	gaugeID, gauge := suite.CreateGauge(isPerpetual, addr, coins, distrTo, startTime2, numEpochsPaidOver)
	return gaugeID, gauge, coins, startTime2
}

// TODO: Delete all usages of this method
func (suite *KeeperTestSuite) SetupNewGauge(isPerpetual bool, coins sdk.Coins) (uint64, *types.Gauge, sdk.Coins, time.Time) {
	return suite.setupNewGaugeWithDuration(isPerpetual, coins, defaultLockDuration)
}

func (suite *KeeperTestSuite) SetupManyLocks(numLocks int, liquidBalance sdk.Coins, coinsPerLock sdk.Coins,
	lockDuration time.Duration) []sdk.AccAddress {
	addrs := make([]sdk.AccAddress, 0, numLocks)
	randPrefix := make([]byte, 8)
	_, _ = rand.Read(randPrefix)

	bal := liquidBalance.Add(coinsPerLock...)
	for i := 0; i < numLocks; i++ {
		addr := suite.setupAddr(i, string(randPrefix), bal)
		_, err := suite.app.LockupKeeper.LockTokens(suite.ctx, addr, coinsPerLock, lockDuration)
		suite.Require().NoError(err)
		addrs = append(addrs, addr)
	}
	return addrs
}

func (suite *KeeperTestSuite) SetupLockAndGauge(isPerpetual bool) (sdk.AccAddress, uint64, sdk.Coins, time.Time) {
	// create a gauge and locks
	lockOwner := sdk.AccAddress([]byte("addr1---------------"))
	suite.LockTokens(lockOwner, sdk.Coins{sdk.NewInt64Coin("lptoken", 10)}, time.Second)

	// create gauge
	gaugeID, _, gaugeCoins, startTime := suite.SetupNewGauge(isPerpetual, sdk.Coins{sdk.NewInt64Coin("stake", 10)})

	return lockOwner, gaugeID, gaugeCoins, startTime
}

func (suite *KeeperTestSuite) createAppendGaugeDetails(tcGaugeDetails gaugeDetailsMap, index int, gaugeID uint64, gaugeCoins sdk.Coins, gaugeStartTime time.Time) gaugeDetails {
	// create gauge details
	curGaugeDetails := gaugeDetails{
		gaugeID:        gaugeID,
		gaugeCoins:     gaugeCoins,
		gaugeStartTime: gaugeStartTime,
	}

	// add gauge details to map for later use
	tcGaugeDetails[index] = gaugeDetails{
		gaugeID:        gaugeID,
		gaugeCoins:     gaugeCoins,
		gaugeStartTime: gaugeStartTime,
	}

	return curGaugeDetails
}

func (suite *KeeperTestSuite) addToCoinsFromGaugesOrInitlize(gaugeCoins sdk.Coins, coinsFromGauges sdk.Coins) sdk.Coins {
	if coinsFromGauges == nil {
		coinsFromGauges = gaugeCoins
	} else {
		coinsFromGauges = coinsFromGauges.Add(gaugeCoins...)
	}

	return coinsFromGauges
}

func (suite *KeeperTestSuite) validateTestcaseGauges(tcGauges []gauge, isPerpetual bool, gaugeDetailsMap gaugeDetailsMap, coinsFromGauges sdk.Coins) (gaugeDetailsMap, sdk.Coins) {
	for index, gauge := range tcGauges {
		gaugeDesc := fmt.Sprintf("gauge index=%v", index)
		suite.Run(gaugeDesc, func() {
			// setup lock and gauge
			gaugeID, gaugeCoins, gaugeStartTime := suite.setupGaugeHelper(gauge.isPerpetual, gauge.isLock)

			// create gauge details
			// add gauge details to map for later use
			suite.createAppendGaugeDetails(gaugeDetailsMap, index, gaugeID, gaugeCoins, gaugeStartTime)

			// validate gauge initlization
			suite.validateGaugeInitlization(index, tcGauges, gaugeDetailsMap)

			// set coinsFromGauges, if nil initialize with 1st gauge coins
			// else add to running total
			coinsFromGauges = suite.addToCoinsFromGaugesOrInitlize(gaugeCoins, coinsFromGauges)

			// check addition if gauge include coinsToAdd
			if gauge.coinsToAdd != nil {
				coinsFromGauges = suite.validateAdditionToGauge(index, tcGauges, gaugeDetailsMap, coinsFromGauges)
			}
		})
	}

	return gaugeDetailsMap, coinsFromGauges
}

func (suite *KeeperTestSuite) setupGaugeHelper(isPerpetual bool, isLock bool) (uint64, sdk.Coins, time.Time) {
	var gaugeID uint64
	var gaugeCoins sdk.Coins
	var gaugeStartTime time.Time

	// setup lock and gauge
	if isLock {
		_, gaugeID, gaugeCoins, gaugeStartTime = suite.SetupLockAndGauge(isPerpetual)
	} else {
		gaugeID, _, gaugeCoins, gaugeStartTime = suite.SetupNewGauge(isPerpetual, sdk.Coins{sdk.NewInt64Coin("stake", 10)})
	}

	return gaugeID, gaugeCoins, gaugeStartTime
}

func (suite *KeeperTestSuite) getExpectedGauge(g gauge, gDetails gaugeDetails) *types.Gauge {
	var numEpochsPaidOver uint64
	if g.isPerpetual {
		numEpochsPaidOver = 1
	} else {
		numEpochsPaidOver = 2
	}

	return &types.Gauge{
		Id:          gDetails.gaugeID,
		IsPerpetual: g.isPerpetual,
		DistributeTo: lockuptypes.QueryCondition{
			LockQueryType: lockuptypes.ByDuration,
			Denom:         "lptoken",
			Duration:      time.Second,
		},
		Coins:             sdk.Coins{sdk.NewInt64Coin("stake", 10)},
		NumEpochsPaidOver: numEpochsPaidOver,
		FilledEpochs:      0,
		DistributedCoins:  sdk.Coins(nil),
		StartTime:         gDetails.gaugeStartTime,
	}
}

func (suite *KeeperTestSuite) validateIncentivesModuleInitialization() {
	coins := suite.app.IncentivesKeeper.GetModuleToDistributeCoins(suite.ctx)
	suite.Require().Equal(coins, sdk.Coins(nil))

	gauges := suite.app.IncentivesKeeper.GetNotFinishedGauges(suite.ctx)
	suite.Require().Len(gauges, 0)
}

func (suite *KeeperTestSuite) validateGaugeInitlization(index int, tcGauges []gauge, tcGaugeDetails gaugeDetailsMap) {
	isPerpetual := tcGauges[index].isPerpetual
	// check perpetual gauges initlization
	if isPerpetual {
		suite.validatePerptualGaugeInitlization(index, tcGauges, tcGaugeDetails)
	} else {
		suite.validateNonPerptualGaugeInitlization(index, tcGauges, tcGaugeDetails)
	}
}

func (suite *KeeperTestSuite) validateNonPerptualGaugeInitlization(index int, tcGauges []gauge, tcGaugeDetails gaugeDetailsMap) {
	gauges := suite.app.IncentivesKeeper.GetNotFinishedGauges(suite.ctx)
	expectedGauge := suite.getExpectedGauge(tcGauges[index], tcGaugeDetails[index])

	suite.Require().Len(gauges, index+1)
	suite.Require().Equal(gauges[index].Id, expectedGauge.Id)
	suite.Require().Equal(gauges[index].Coins, expectedGauge.Coins)
	suite.Require().Equal(gauges[index].NumEpochsPaidOver, expectedGauge.NumEpochsPaidOver)
	suite.Require().Equal(gauges[index].FilledEpochs, expectedGauge.FilledEpochs)
	suite.Require().Equal(gauges[index].DistributedCoins, expectedGauge.DistributedCoins)
	suite.Require().Equal(gauges[index].StartTime.Unix(), expectedGauge.StartTime.Unix())
}

func (suite *KeeperTestSuite) validatePerptualGaugeInitlization(index int, tcGauges []gauge, tcGaugeDetails gaugeDetailsMap) {
	gauge := tcGauges[index]
	curGaugeDetails := tcGaugeDetails[index]
	gauges := suite.app.IncentivesKeeper.GetNotFinishedGauges(suite.ctx)
	suite.Require().Len(gauges, 1)

	expectedGauge := suite.getExpectedGauge(gauge, curGaugeDetails)
	suite.Require().Equal(gauges[0].String(), expectedGauge.String())
}

func (suite *KeeperTestSuite) validateAdditionToGauge(gaugeIndex int, tcGauges []gauge, tcGaugeDetails gaugeDetailsMap, coinsFromGauges sdk.Coins) sdk.Coins {
	addCoins := tcGauges[gaugeIndex].coinsToAdd
	// validate addition
	suite.AddToGauge(addCoins, tcGaugeDetails[gaugeIndex].gaugeID)
	coinsFromGauges = coinsFromGauges.Add(addCoins...)
	currentCtxCoins := suite.app.IncentivesKeeper.GetModuleToDistributeCoins(suite.ctx)
	suite.Require().Equal(currentCtxCoins, coinsFromGauges)

	return coinsFromGauges
}

func (suite *KeeperTestSuite) validateUpcomingGaugesBeforeDistribution(tcGauges []gauge) {
	gauges := suite.app.IncentivesKeeper.GetUpcomingGauges(suite.ctx)
	suite.Require().Len(gauges, len(tcGauges))
}

func (suite *KeeperTestSuite) validateDistribution(tcGauges []gauge, tcGaugeDetails gaugeDetailsMap, isPerpetual bool, coinsFromGauges sdk.Coins) {
	startingGauge := tcGauges[0]
	startingGaugeDetails := tcGaugeDetails[0]
	suite.ctx = suite.ctx.WithBlockTime(startingGaugeDetails.gaugeStartTime)
	gauge, err := suite.app.IncentivesKeeper.GetGaugeByID(suite.ctx, startingGaugeDetails.gaugeID)
	suite.Require().NoError(err)
	err = suite.app.IncentivesKeeper.BeginDistribution(suite.ctx, *gauge)
	suite.Require().NoError(err)

	// distribute coins to stakers
	distrCoins, err := suite.app.IncentivesKeeper.Distribute(suite.ctx, []types.Gauge{*gauge})
	suite.Require().NoError(err)

	// since it's perpetual distribute everything on single distribution
	if isPerpetual {
		suite.Require().Equal(distrCoins, sdk.Coins(nil))
	}

	if !startingGauge.isLock && !startingGauge.isPerpetual {
		suite.Require().Equal(distrCoins, sdk.Coins(nil))
	} else if !isPerpetual {
		suite.Require().Equal(distrCoins, sdk.Coins{sdk.NewInt64Coin("stake", 105)})
	}

	// check gauge changes after distribution
	coins := suite.app.IncentivesKeeper.GetModuleToDistributeCoins(suite.ctx)
	suite.Require().Equal(coins, coinsFromGauges.Sub(distrCoins))

	gauges := suite.app.IncentivesKeeper.GetNotFinishedGauges(suite.ctx)
	suite.Require().Len(gauges, len(tcGauges))

	if isPerpetual {
		suite.Require().Equal(gauges[0].String(), suite.getExpectedGauge(tcGauges[0], startingGaugeDetails).String())
	}
}
