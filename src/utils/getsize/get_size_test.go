package getsize

import "testing"

func TestGetSize(t *testing.T) {
	tests := []struct {
		name     string
		input    int64
		wantSize float64
		wantUnit string
	}{
		{name: "bytes", input: 512, wantSize: 512, wantUnit: "B"},
		{name: "kilobyte boundary", input: KB, wantSize: 1, wantUnit: "KB"},
		{name: "megabyte boundary", input: MB, wantSize: 1, wantUnit: "MB"},
		{name: "gigabyte boundary", input: GB, wantSize: 1, wantUnit: "GB"},
		{name: "terabyte boundary", input: TB, wantSize: 1, wantUnit: "TB"},
		{name: "fractional megabytes", input: 5 * MB / 2, wantSize: 2.5, wantUnit: "MB"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			gotSize, gotUnit := GetSize(test.input)
			if gotSize != test.wantSize || gotUnit != test.wantUnit {
				t.Fatalf(
					"GetSize(%d) = (%v, %q); want (%v, %q)",
					test.input,
					gotSize,
					gotUnit,
					test.wantSize,
					test.wantUnit,
				)
			}
		})
	}
}
