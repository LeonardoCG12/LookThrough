package gethardware

import (
	"runtime"
	"testing"
)

func TestCalculateDynamicWorkerLimit(t *testing.T) {
	original := runtime.GOMAXPROCS(0)
	t.Cleanup(func() {
		runtime.GOMAXPROCS(original)
	})

	tests := []struct {
		parallelism int
		want        int
	}{
		{parallelism: 1, want: 4},
		{parallelism: 2, want: 4},
		{parallelism: 4, want: 8},
		{parallelism: 8, want: 16},
		{parallelism: 12, want: 24},
		{parallelism: 16, want: 32},
		{parallelism: 64, want: 32},
	}

	for _, test := range tests {
		runtime.GOMAXPROCS(test.parallelism)

		if got := CalculateDynamicWorkerLimit(); got != test.want {
			t.Fatalf(
				"CalculateDynamicWorkerLimit() with GOMAXPROCS=%d returned %d; want %d",
				test.parallelism,
				got,
				test.want,
			)
		}
	}
}

func TestResolveWorkerLimit(t *testing.T) {
	original := runtime.GOMAXPROCS(0)
	runtime.GOMAXPROCS(8)
	t.Cleanup(func() {
		runtime.GOMAXPROCS(original)
	})

	tests := []struct {
		name      string
		requested int
		want      int
		wantError bool
	}{
		{name: "automatic", requested: 0, want: 16},
		{name: "minimum manual", requested: 1, want: 1},
		{name: "ordinary manual", requested: 32, want: 32},
		{name: "maximum manual", requested: MaximumManualWorkers, want: MaximumManualWorkers},
		{name: "negative", requested: -1, wantError: true},
		{name: "above maximum", requested: MaximumManualWorkers + 1, wantError: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := ResolveWorkerLimit(test.requested)

			if test.wantError {
				if err == nil {
					t.Fatalf("ResolveWorkerLimit(%d) returned no error", test.requested)
				}
				return
			}

			if err != nil {
				t.Fatalf("ResolveWorkerLimit(%d) returned error: %v", test.requested, err)
			}

			if got != test.want {
				t.Fatalf("ResolveWorkerLimit(%d) returned %d; want %d", test.requested, got, test.want)
			}
		})
	}
}
