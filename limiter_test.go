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
	now := time.Now()
	tick := time.Second
	window := 10 * tick
	max := decimal.NewFromInt(10)
	lb := leakybucket.NewLeakyBucket(tick, window, max)
	assert.Equal(t, max, lb.Peek(now))
	assert.NoError(t, lb.Allow(now, decimal.NewFromInt(4)))
	assert.True(t, lb.Peek(now).Equal(decimal.NewFromInt(6)))
	assert.NoError(t, lb.Allow(now, decimal.NewFromInt(4)))
	assert.Error(t, lb.Allow(now, decimal.NewFromInt(3)))
	now = now.Add(tick)
	assert.NoError(t, lb.Allow(now, decimal.NewFromInt(3)))
}

func TestPartialCreditRetain(t *testing.T) {
	tick := time.Second
	window := 5 * tick
	max := int64(5)
	now := time.Now()
	lb := leakybucket.NewLeakyBucket(tick, window, decimal.NewFromInt(max))
	// Drain available
	err := lb.Allow(now, decimal.NewFromInt(5))
	assert.NoError(t, err)
	// Check available at 1.5 Ticks
	now = now.Add(time.Second + (500 * time.Millisecond))
	// Should have one credit, and a half tick to go for next credit
	assert.True(t, lb.Peek(now).Equal(decimal.NewFromInt(1)))
	// Sending 1 should succeed and retain the half tick
	err = lb.Allow(now, decimal.NewFromInt(1))
	assert.NoError(t, err)
	// Add a half tick
	now = now.Add(500 * time.Millisecond)
	// Should have one
	avail := lb.Peek(now)
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
			name: "Negative Allow attempt",
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
			now := time.Now()
			lb := leakybucket.NewLeakyBucket(tick, window, decimal.NewFromInt(max))
			last := len(tt.args.tick) - 1
			for i := 0; i <= last; i++ {
				now = now.Add(tt.args.tick[i])
				err := lb.Allow(now, decimal.NewFromInt(tt.args.amount[i]))
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
