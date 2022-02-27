package leakybucket

import "time"

var (
	_ Clock = (*SystemClock)(nil)
	_ Clock = (*MockClock)(nil)
)

type Clock interface {
	Now() time.Time
}

type SystemClock struct{}

func (SystemClock) Now() time.Time { return time.Now() }

type MockClock struct {
	Clock time.Time
}

func NewMockClock() *MockClock {
	return &MockClock{
		Clock: time.Now(),
	}
}

func (m *MockClock) Now() time.Time {
	return m.Clock
}
