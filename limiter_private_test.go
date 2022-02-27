package leakybucket

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func Test_credits(t *testing.T) {
	type args struct {
		from          time.Time
		to            time.Time
		replenishRate time.Duration
	}
	now := time.Now()
	tests := []struct {
		name          string
		args          args
		wantCredits   int64
		wantRemainder int64
	}{
		{
			name: "Zero",
			args: args{
				from:          now,
				to:            now,
				replenishRate: time.Second,
			},
			wantCredits:   0,
			wantRemainder: 0,
		},
		{
			name: "Exactly One",
			args: args{
				from:          now,
				to:            now.Add(time.Second),
				replenishRate: time.Second,
			},
			wantCredits:   1,
			wantRemainder: 0,
		},
		{
			name: "One and 100",
			args: args{
				from:          now,
				to:            now.Add(time.Second + 100*time.Nanosecond),
				replenishRate: time.Second,
			},
			wantCredits:   1,
			wantRemainder: 100,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotCredits, gotRemainder := credits(tt.args.replenishRate, tt.args.from, tt.args.to)
			assert.Equalf(t, tt.wantCredits, gotCredits, "credits(%v, %v)", tt.args.from, tt.args.to)
			assert.Equalf(t, tt.wantRemainder, gotRemainder, "credits(%v, %v)", tt.args.from, tt.args.to)
		})
	}
}
