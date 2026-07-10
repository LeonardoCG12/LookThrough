package variables

type FileHash struct {
	Name string
	Hash string
}

type LookThroughVars struct {
	FileCount      int
	HashCount      int
	HashList       []FileHash
	HashListAll    []FileHash
	HashMap        map[string]bool
	NameMap        map[string]bool
	Mem            map[string]int
	MyPath         string
	NewPath        string
	Num            string
	Separator      string
	SizeCount      int64
	TotalSizeCount int64
}
