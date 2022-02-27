package leakybucket_test

import (
	"github.com/stretchr/testify/assert"
	"leakybucket"
	"testing"
	"time"
)

func TestRealClock_Now(t *testing.T) {
	clock := leakybucket.SystemClock{}
	now := time.Now()

	systemNow := clock.Now()
	assert.GreaterOrEqual(t, systemNow.Unix(), now.Unix())
}

func TestMockClock_Now(t *testing.T) {
	now := time.Now()
	mockClock := leakybucket.NewMockClock()

	assert.GreaterOrEqual(t, mockClock.Now().Unix(), now.Unix())

	hourAgo := now.Add(-time.Hour)
	mockClock.Clock = hourAgo
	assert.Equal(t, mockClock.Now().Unix(), hourAgo.Unix())
}
