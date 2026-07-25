package variables

import "crypto/sha256"

type Digest [sha256.Size]byte

type FileHash struct {
	Name string
	Hash Digest
}

type LookThroughVars struct {
	FileCount             int64
	HashCount             int64
	HashList              []FileHash
	HashListAll           []FileHash
	HashMap               map[Digest]struct{}
	NameMap               map[string]struct{}
	Mem                   map[string]int
	MyPath                string
	IgnoreDirectoryErrors bool
	NewPath               string
	SafeCopy              bool
	ShowProgressBar       bool
	SizeCount             int64
	SortFile              bool
	TotalSizeCount        int64
	VerifyFiles           bool
	Workers               int
}
