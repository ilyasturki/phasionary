package app

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"phasionary/internal/config"
)

func currentOptionValue(t *testing.T, m *model, i int) string {
	t.Helper()
	return optionSpecs[i].current(m.deps.CfgManager.Get())
}

func TestCycleOption_ForwardWrapsRound(t *testing.T) {
	m := newTestModel(t, sampleProject())
	m.ui.Options.selectedOption = 1 // Priority Color

	require.Equal(t, config.PriorityColorFull, currentOptionValue(t, m, 1))
	m.cycleSelectedOption(1)
	assert.Equal(t, config.PriorityColorIcon, currentOptionValue(t, m, 1))
	m.cycleSelectedOption(1)
	assert.Equal(t, config.PriorityColorNone, currentOptionValue(t, m, 1))
	m.cycleSelectedOption(1)
	assert.Equal(t, config.PriorityColorFull, currentOptionValue(t, m, 1))
}

func TestCycleOption_BackwardIsTheInverse(t *testing.T) {
	m := newTestModel(t, sampleProject())
	m.ui.Options.selectedOption = 1

	m.cycleSelectedOption(-1)
	assert.Equal(t, config.PriorityColorNone, currentOptionValue(t, m, 1))
	m.cycleSelectedOption(1)
	assert.Equal(t, config.PriorityColorFull, currentOptionValue(t, m, 1))
}

// Every spec must be able to find its own current value among its values, or
// the dialog would highlight nothing and h/l would always land on the first.
func TestOptionSpecs_CurrentValueIsListed(t *testing.T) {
	m := newTestModel(t, sampleProject())
	for i, spec := range optionSpecs {
		cur := spec.current(m.deps.CfgManager.Get())
		keys := make([]string, len(spec.values))
		for j, v := range spec.values {
			keys[j] = v.key
		}
		assert.Containsf(t, keys, cur, "option %d (%s)", i, spec.name)
	}
}

func TestOptionSpecs_EveryValueRoundTrips(t *testing.T) {
	for i, spec := range optionSpecs {
		for _, v := range spec.values {
			m := newTestModel(t, sampleProject())
			m.ui.Options.selectedOption = i
			spec.apply(m, v.key)
			assert.Equalf(t, v.key, spec.current(m.deps.CfgManager.Get()), "option %s value %s", spec.name, v.key)
		}
	}
}
