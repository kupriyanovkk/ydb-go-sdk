package operation

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/ydb-platform/ydb-go-sdk/v3/internal/timeutil"
)

func TestOperationParams(t *testing.T) {
	for _, test := range [...]struct {
		name string

		ctxTimeout time.Duration

		opTimeout time.Duration
		opCancel  time.Duration
		opMode    OperationMode

		exp Params
	}{
		{
			name: "nothing",
		},
		{
			name:       "mode: unknown, deadline timeout",
			ctxTimeout: time.Second,
			exp:        Params{},
		},
		{
			name:       "mode: sync, deadline timeout applied to operation timeout",
			ctxTimeout: time.Second,
			opMode:     OperationModeSync,
			exp: Params{
				Timeout: time.Second,
				Mode:    OperationModeSync,
			},
		},
		{
			name:       "mode: async, deadline timeout not applied to operation timeout",
			ctxTimeout: time.Second,
			opMode:     OperationModeAsync,
			exp: Params{
				Mode: OperationModeAsync,
			},
		},
		{
			name:       "mode: unknown, deadline timeout not override operation timeout",
			ctxTimeout: time.Second,
			opTimeout:  time.Hour,
			exp: Params{
				Timeout: time.Hour,
			},
		},
		{
			name:       "mode: sync, deadline timeout override operation timeout",
			ctxTimeout: time.Second,
			opMode:     OperationModeSync,
			opTimeout:  time.Hour,
			exp: Params{
				Timeout: time.Second,
				Mode:    OperationModeSync,
			},
		},
		{
			name:       "mode: async, deadline timeout not override operation timeout",
			ctxTimeout: time.Second,
			opMode:     OperationModeAsync,
			opTimeout:  time.Hour,
			exp: Params{
				Timeout: time.Hour,
				Mode:    OperationModeAsync,
			},
		},
		{
			name:       "mode: unknown, cancel after timeout",
			ctxTimeout: time.Second,
			opCancel:   time.Hour,
			exp: Params{
				CancelAfter: time.Hour,
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			_, cleanupNow := timeutil.StubTestHookTimeNow(time.Unix(0, 0))
			defer cleanupNow()

			ctx := context.Background()
			if t := test.opTimeout; t > 0 {
				ctx = WithOperationTimeout(ctx, t)
			}
			if t := test.opCancel; t > 0 {
				ctx = WithOperationCancelAfter(ctx, t)
			}
			if m := test.opMode; m != 0 {
				ctx = WithOperationMode(ctx, m)
			}
			if t := test.ctxTimeout; t > 0 {
				var cancel context.CancelFunc
				ctx, cancel = context.WithDeadline(ctx, timeutil.Now().Add(t))
				defer cancel()
			}

			act := ContextParams(ctx)

			if exp := test.exp; !reflect.DeepEqual(act, exp) {
				t.Fatalf(
					"unexpected operation parameters: %v; want %v",
					act, exp,
				)
			}
		})
	}
}
