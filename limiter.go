package leakybucket

import (
	"fmt"
	"github.com/shopspring/decimal"
	"sync"
	"time"
)

type LeakyBucket struct {
	m               sync.Mutex
	clock           Clock
	lastAdjusted    time.Time
	replenishRate   time.Duration
	capacity        decimal.Decimal
	replenishAmount decimal.Decimal
	spent           decimal.Decimal
}

func NewLeakyBucket(clock Clock, replenishRate time.Duration, period time.Duration, capacity decimal.Decimal) *LeakyBucket {
	periodDecimal := decimal.NewFromInt(int64(period))
	replenishRateDecimal := decimal.NewFromInt(int64(replenishRate))
	replenishAmount := capacity.Div(periodDecimal.Div(replenishRateDecimal))
	return &LeakyBucket{
		clock:           clock,
		replenishRate:   replenishRate,
		capacity:        capacity,
		replenishAmount: replenishAmount,
	}
}

// Available returns the *currently* available amount. The returned amount is *not* guaranteed to be available to later
// spend if this LeakyBucket is being shared.
func (lb *LeakyBucket) Available() decimal.Decimal {
	lb.m.Lock()
	defer lb.m.Unlock()
	at := lb.clock.Now()
	avail, _ := lb.lockedAvailable(at)
	return avail
}

// Spend attempts to spend a given amount. If the amount is available it will be spent, if not an error is returned.
func (lb *LeakyBucket) Spend(amount decimal.Decimal) error {
	if amount.IsNegative() {
		return fmt.Errorf("can not spend negative amounts")
	}
	lb.m.Lock()
	defer lb.m.Unlock()
	at := lb.clock.Now()
	var avail decimal.Decimal
	avail, at = lb.lockedAvailable(at)
	if amount.GreaterThan(avail) {
		return fmt.Errorf("%s greater than available %s", amount, avail)
	}
	lb.lockedSpend(amount, at)
	return nil
}

// lockedAvailable calculates the available amount, applying any credits due. This functions assumes the lock is
// already acquired and the calculations can be safely performed.
func (lb *LeakyBucket) lockedAvailable(at time.Time) (decimal.Decimal, time.Time) {
	if lb.spent.GreaterThan(decimal.Zero) {
		credits, remainder := credits(lb.replenishRate, lb.lastAdjusted, at)
		if credits > 0 {
			at = at.Add(-time.Duration(remainder) * time.Nanosecond)
			cd := decimal.NewFromInt(credits)
			lb.lockedSpend(cd.Mul(lb.replenishAmount).Neg(), at)
		} else {
			at = lb.lastAdjusted
		}
	}
	return lb.capacity.Sub(lb.spent), at
}

// lockedSpend adjusts the spend, up or down, by a given amount. This function assumes the lock is
// already acquired and the calculations can be safely performed.
func (lb *LeakyBucket) lockedSpend(amount decimal.Decimal, at time.Time) {
	adjusted := lb.spent.Add(amount)
	if adjusted.LessThan(decimal.Zero) {
		adjusted = decimal.Zero
	}
	lb.spent = adjusted
	lb.lastAdjusted = at
}

func credits(replenishRate time.Duration, from time.Time, to time.Time) (credits int64, remainder int64) {
	elapsed := to.Sub(from)
	credits = elapsed.Nanoseconds() / replenishRate.Nanoseconds()
	remainder = elapsed.Nanoseconds() % replenishRate.Nanoseconds()
	return credits, remainder
}
