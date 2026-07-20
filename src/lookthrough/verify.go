package lookthrough

func (l *LookThrough) verifyFiles() bool {
	hashSet := make(map[string]bool)

	for _, fh := range l.Vars.HashList {
		hashSet[fh.Hash] = true
	}

	for _, all := range l.Vars.HashListAll {
		if !hashSet[all.Hash] {
			return false
		}
	}

	return true
}
