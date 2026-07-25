package lookthrough

import "github.com/LeonardoCG12/LookThrough/src/utils/variables"

func (l *LookThrough) verifyFiles() bool {
	hashSet := make(map[variables.Digest]struct{}, len(l.Vars.HashList))

	for _, fileHash := range l.Vars.HashList {
		hashSet[fileHash.Hash] = struct{}{}
	}

	for _, scannedFile := range l.Vars.HashListAll {
		if _, exists := hashSet[scannedFile.Hash]; !exists {
			return false
		}
	}

	return true
}
