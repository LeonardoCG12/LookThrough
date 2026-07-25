package lookthrough

import (
	"testing"

	"github.com/LeonardoCG12/LookThrough/src/utils/variables"
)

func TestVerifyFiles(t *testing.T) {
	first := variables.Digest{1}
	second := variables.Digest{2}
	missing := variables.Digest{3}

	tests := []struct {
		name        string
		hashList    []variables.FileHash
		hashListAll []variables.FileHash
		want        bool
	}{
		{
			name: "empty lists",
			want: true,
		},
		{
			name: "all scanned hashes are represented",
			hashList: []variables.FileHash{
				{Name: "first.txt", Hash: first},
				{Name: "second.txt", Hash: second},
			},
			hashListAll: []variables.FileHash{
				{Name: "first.txt", Hash: first},
				{Name: "duplicate-first.txt", Hash: first},
				{Name: "second.txt", Hash: second},
			},
			want: true,
		},
		{
			name: "scanned hash is missing",
			hashList: []variables.FileHash{
				{Name: "first.txt", Hash: first},
			},
			hashListAll: []variables.FileHash{
				{Name: "first.txt", Hash: first},
				{Name: "missing.txt", Hash: missing},
			},
			want: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			lookThrough := NewLookThrough(variables.LookThroughVars{
				HashList:    test.hashList,
				HashListAll: test.hashListAll,
				Workers:     1,
			})

			if got := lookThrough.verifyFiles(); got != test.want {
				t.Fatalf("verifyFiles() = %t; want %t", got, test.want)
			}
		})
	}
}
