package lookthrough

import (
	"fmt"
	"os"
	"sync"

	"github.com/LeonardoCG12/LookThrough/src/utils/gethardware"
	"github.com/LeonardoCG12/LookThrough/src/utils/variables"
)

type directoryCreation struct {
	done chan struct{}
	err  error
}

type LookThrough struct {
	Vars variables.LookThroughVars

	mu                 sync.Mutex
	workers            int
	createdDirs        sync.Map
	scanMu             sync.Mutex
	skippedDirectories map[string]string
}

func NewLookThrough(vars variables.LookThroughVars) *LookThrough {
	if vars.Mem == nil {
		vars.Mem = make(map[string]int)
	}
	if vars.HashMap == nil {
		vars.HashMap = make(map[variables.Digest]struct{})
	}
	if vars.NameMap == nil {
		vars.NameMap = make(map[string]struct{})
	}

	workers := vars.Workers
	if workers <= 0 {
		workers = gethardware.CalculateDynamicWorkerLimit()
	}
	vars.Workers = workers

	return &LookThrough{
		Vars:               vars,
		workers:            workers,
		skippedDirectories: make(map[string]string),
	}
}

func (l *LookThrough) ensureDirectory(path string) error {
	creation := &directoryCreation{done: make(chan struct{})}
	actual, loaded := l.createdDirs.LoadOrStore(path, creation)

	if loaded {
		existing := actual.(*directoryCreation)
		<-existing.done
		return existing.err
	}

	creation.err = os.MkdirAll(path, 0755)
	close(creation.done)

	if creation.err != nil {
		l.createdDirs.Delete(path)
		return fmt.Errorf("failed to create directory '%s': %w", path, creation.err)
	}

	return nil
}
