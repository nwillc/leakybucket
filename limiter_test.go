package leakybucket_test

import (
	"fmt"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"leakybucket"
	"testing"
	"time"
)

func TestNewLeakyBucket(t *testing.T) {
	clock := leakybucket.NewMockClock()

	tick := time.Second
	window := 10 * tick
	max := decimal.NewFromInt(10)
	lb := leakybucket.NewLeakyBucket(clock, tick, window, max)
	assert.Equal(t, max, lb.Available())
	assert.NoError(t, lb.Spend(decimal.NewFromInt(4)))
	assert.True(t, lb.Available().Equal(decimal.NewFromInt(6)))
	assert.NoError(t, lb.Spend(decimal.NewFromInt(4)))
	assert.Error(t, lb.Spend(decimal.NewFromInt(3)))
	clock.Clock = clock.Clock.Add(tick)
	assert.NoError(t, lb.Spend(decimal.NewFromInt(3)))
}

func TestPartialCreditRetain(t *testing.T) {
	tick := time.Second
	window := 5 * tick
	max := int64(5)
	clock := leakybucket.NewMockClock()
	lb := leakybucket.NewLeakyBucket(clock, tick, window, decimal.NewFromInt(max))
	// Drain available
	err := lb.Spend(decimal.NewFromInt(5))
	assert.NoError(t, err)
	// Check available at 1.5 ticks
	clock.Clock = clock.Clock.Add(time.Second + (500 * time.Millisecond))
	// Should have one credit, and a half tick to go for next credit
	assert.True(t, lb.Available().Equal(decimal.NewFromInt(1)))
	// Sending 1 should succeed and retain the half tick
	err = lb.Spend(decimal.NewFromInt(1))
	assert.NoError(t, err)
	// Add a half tick
	clock.Clock = clock.Clock.Add(500 * time.Millisecond)
	// Should have one
	avail := lb.Available()
	fmt.Println(avail)
	assert.True(t, avail.Equal(decimal.NewFromInt(1)))
}

func TestScenarios(t *testing.T) {
	tick := time.Second
	window := 5 * tick
	max := int64(10)

	type args struct {
		tick   []time.Duration
		amount []int64
	}
	tests := []struct {
		name     string
		args     args
		hasError bool
	}{
		{
			name: "Negative Spend attempt",
			args: args{
				tick:   []time.Duration{0},
				amount: []int64{-1},
			},
			hasError: true,
		},
		{
			name: "Single exceeds",
			args: args{
				tick:   []time.Duration{0},
				amount: []int64{max + 1},
			},
			hasError: true,
		},
		{
			name: "Multiple exceeds",
			args: args{
				tick:   []time.Duration{0, 0},
				amount: []int64{5, 6},
			},
			hasError: true,
		},
		{
			name: "Exceeds with one tick",
			args: args{
				tick:   []time.Duration{0, tick},
				amount: []int64{3, 10},
			},
			hasError: true,
		},
		{
			name: "Single succeeds",
			args: args{
				tick:   []time.Duration{0},
				amount: []int64{max},
			},
			hasError: false,
		},
		{
			name: "Multiple exceeds",
			args: args{
				tick:   []time.Duration{0, 0},
				amount: []int64{5, 5},
			},
			hasError: false,
		},
		{
			name: "Succeeds with one tick",
			args: args{
				tick:   []time.Duration{0, time.Second},
				amount: []int64{2, max},
			},
			hasError: false,
		},
		{
			name: "Simple single",
			args: args{
				tick:   []time.Duration{time.Second, time.Second, time.Second},
				amount: []int64{3, 3, 3},
			},
			hasError: false,
		},
		{
			name: "Simple Multiple",
			args: args{
				tick:   []time.Duration{time.Second, time.Second, time.Second, time.Second, time.Second, time.Second},
				amount: []int64{2, 2, 2, 2, 2, 2},
			},
			hasError: false,
		},
		{
			name: "Simple longer delay",
			args: args{
				tick:   []time.Duration{time.Second, 4 * time.Second},
				amount: []int64{2, 8},
			},
			hasError: false,
		},

		{
			name: "longer delay exceeds",
			args: args{
				tick:   []time.Duration{time.Second, 4 * time.Second},
				amount: []int64{2, 9},
			},
			hasError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clock := leakybucket.NewMockClock()
			lb := leakybucket.NewLeakyBucket(clock, tick, window, decimal.NewFromInt(max))
			last := len(tt.args.tick) - 1
			for i := 0; i <= last; i++ {
				clock.Clock = clock.Clock.Add(tt.args.tick[i])
				err := lb.Spend(decimal.NewFromInt(tt.args.amount[i]))
				if i != last {
					assert.NoError(t, err)
				} else {
					if tt.hasError {
						fmt.Println(err)
						assert.Error(t, err)
					} else {
						assert.NoError(t, err)
					}
				}
			}
		})
	}
}
