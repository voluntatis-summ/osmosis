package simulation_test

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/osmosis-labs/osmosis/v7/x/mint/simulation"
)

func TestParamChanges(t *testing.T) {
	s := rand.NewSource(1)
	r := rand.New(s)

	expected := []struct {
		composedKey string
		key         string
		simValue    string
		subspace    string
	}{
		// {"mint/MaxRewardPerEpoch", "MaxRewardPerEpoch", "\"0.200000000000000000\"", "mint"},
		// {"mint/MinRewardPerEpoch", "MinRewardPerEpoch", "\"0.070000000000000000\"", "mint"},
		// Leaving as sample code

		// TODO: Test proper ParamChanges
	}

	paramChanges := simulation.ParamChanges(r)
	// require.Len(t, paramChanges, 2)

	for i, p := range paramChanges {
		require.Equal(t, expected[i].composedKey, p.ComposedKey())
		require.Equal(t, expected[i].key, p.Key())
		require.Equal(t, expected[i].simValue, p.SimValue()(r))
		require.Equal(t, expected[i].subspace, p.Subspace())
	}

}
