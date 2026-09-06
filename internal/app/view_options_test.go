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

func TestCycleOption_WrapsBothWays(t *testing.T) {
	m := newTestModel(t, sampleProject())
	m.ui.Options.selectedOption = 1 // Priority Color

	require.Equal(t, config.PriorityColorFull, currentOptionValue(t, m, 1))
	for _, want := range []string{config.PriorityColorIcon, config.PriorityColorNone, config.PriorityColorFull} {
		m.cycleSelectedOption(1)
		assert.Equal(t, want, currentOptionValue(t, m, 1))
	}
	m.cycleSelectedOption(-1)
	assert.Equal(t, config.PriorityColorNone, currentOptionValue(t, m, 1), "h is l's inverse, and wraps too")
}

func TestOptionSpecs_EveryValueRoundTrips(t *testing.T) {
	m := newTestModel(t, sampleProject())
	for _, spec := range optionSpecs {
		keys := make([]string, len(spec.values))
		for j, v := range spec.values {
			keys[j] = v.key
		}
		// Or the dialog would highlight nothing and h/l would land on the first.
		assert.Containsf(t, keys, spec.current(m.deps.CfgManager.Get()), "option %s", spec.name)

		for _, v := range spec.values {
			spec.apply(m, v.key)
			assert.Equalf(t, v.key, spec.current(m.deps.CfgManager.Get()), "option %s value %s", spec.name, v.key)
		}
	}
}
