package leakybucket

import (
	"fmt"
	"github.com/shopspring/decimal"
	"sync"
	"time"
)

// LeakyBucket struct represents a Leaky Bucket based rate limiter.
type LeakyBucket struct {
	m                   sync.Mutex
	lastAdjusted        time.Time
	replenishmentRate   time.Duration
	limit               decimal.Decimal
	replenishmentAmount decimal.Decimal
	spent               decimal.Decimal
}

// NewLeakyBucket initializes a new leaky bucket with the given limit, that replenished by replenishmentAmount each period.
func NewLeakyBucket(replenishmentRate time.Duration, period time.Duration, limit decimal.Decimal) *LeakyBucket {
	periodDecimal := decimal.NewFromInt(int64(period))
	replenishRateDecimal := decimal.NewFromInt(int64(replenishmentRate))
	replenishAmount := limit.Div(periodDecimal.Div(replenishRateDecimal))
	return &LeakyBucket{
		replenishmentRate:   replenishmentRate,
		limit:               limit,
		replenishmentAmount: replenishAmount,
	}
}

// Peek returns the *currently* available amount. There is no guarantee the returned amount will be available later.
func (lb *LeakyBucket) Peek(at time.Time) decimal.Decimal {
	lb.m.Lock()
	defer lb.m.Unlock()
	avail, _ := lb.lockedPeek(at)
	return avail
}

// Allow attempts to spend a given amount. If the amount is available it will be spent, if not an error is returned.
func (lb *LeakyBucket) Allow(at time.Time, amount decimal.Decimal) error {
	if amount.IsNegative() {
		return fmt.Errorf("can not spend negative amounts")
	}
	lb.m.Lock()
	defer lb.m.Unlock()
	var avail decimal.Decimal
	avail, at = lb.lockedPeek(at)
	if amount.GreaterThan(avail) {
		return fmt.Errorf("%s greater than available %s", amount, avail)
	}
	lb.lockedSpend(amount, at)
	return nil
}

// lockedPeek calculates the available amount, applying any Ticks due. This functions assumes the lock is
// already acquired and the calculations can be safely performed.
func (lb *LeakyBucket) lockedPeek(at time.Time) (decimal.Decimal, time.Time) {
	if lb.spent.GreaterThan(decimal.Zero) {
		ticks, remainder := ticks(lb.replenishmentRate, lb.lastAdjusted, at)
		if ticks > 0 {
			at = at.Add(-time.Duration(remainder) * time.Nanosecond)
			lb.lockedSpend(decimal.NewFromInt(ticks).Mul(lb.replenishmentAmount).Neg(), at)
		} else {
			at = lb.lastAdjusted
		}
	}
	return lb.limit.Sub(lb.spent), at
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

// ticks of duration between from and to.
func ticks(tick time.Duration, from time.Time, to time.Time) (credits int64, remainder int64) {
	elapsed := to.Sub(from)
	credits = elapsed.Nanoseconds() / tick.Nanoseconds()
	remainder = elapsed.Nanoseconds() % tick.Nanoseconds()
	return credits, remainder
}
