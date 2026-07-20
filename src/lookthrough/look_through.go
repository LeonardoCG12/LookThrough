package lookthrough

import (
	"sync"

	"github.com/LeonardoCG12/LookThrough/src/utils/gethardware"
	"github.com/LeonardoCG12/LookThrough/src/utils/variables"
)

type LookThrough struct {
	Vars      variables.LookThroughVars
	mu        sync.Mutex
	semaphore chan struct{}
}

func NewLookThrough(vars variables.LookThroughVars) *LookThrough {
	if vars.Mem == nil {
		vars.Mem = make(map[string]int)
	}
	if vars.HashMap == nil {
		vars.HashMap = make(map[string]bool)
	}
	if vars.NameMap == nil {
		vars.NameMap = make(map[string]bool)
	}

	vars.HashList = []variables.FileHash{}
	vars.HashListAll = []variables.FileHash{}

	dynamicLimit := gethardware.CalculateDynamicLimit()

	return &LookThrough{
		Vars:      vars,
		semaphore: make(chan struct{}, dynamicLimit),
	}
}
