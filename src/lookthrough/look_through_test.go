package lookthrough

import (
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/LeonardoCG12/LookThrough/src/utils/variables"
)

func TestNewLookThroughInitializesState(t *testing.T) {
	lookThrough := NewLookThrough(variables.LookThroughVars{Workers: 3})

	if lookThrough.Vars.Mem == nil {
		t.Fatal("expected Mem to be initialized")
	}
	if lookThrough.Vars.HashMap == nil {
		t.Fatal("expected HashMap to be initialized")
	}
	if lookThrough.Vars.NameMap == nil {
		t.Fatal("expected NameMap to be initialized")
	}
	if lookThrough.workers != 3 {
		t.Fatalf("workers = %d; want 3", lookThrough.workers)
	}
	if lookThrough.Vars.Workers != 3 {
		t.Fatalf("Vars.Workers = %d; want 3", lookThrough.Vars.Workers)
	}
}

func TestNewLookThroughPreservesProvidedMaps(t *testing.T) {
	mem := map[string]int{"file.txt": 2}
	hashMap := map[variables.Digest]struct{}{{1}: {}}
	nameMap := map[string]struct{}{"file.txt": {}}

	lookThrough := NewLookThrough(variables.LookThroughVars{
		Mem:     mem,
		HashMap: hashMap,
		NameMap: nameMap,
		Workers: 1,
	})

	if lookThrough.Vars.Mem["file.txt"] != 2 {
		t.Fatal("expected provided Mem contents to be preserved")
	}
	if _, exists := lookThrough.Vars.HashMap[variables.Digest{1}]; !exists {
		t.Fatal("expected provided HashMap contents to be preserved")
	}
	if _, exists := lookThrough.Vars.NameMap["file.txt"]; !exists {
		t.Fatal("expected provided NameMap contents to be preserved")
	}
}

func TestEnsureDirectoryCoordinatesConcurrentCreation(t *testing.T) {
	lookThrough := NewLookThrough(variables.LookThroughVars{Workers: 4})
	target := filepath.Join(t.TempDir(), "nested", "destination")

	const callers = 32
	var waitGroup sync.WaitGroup
	errors := make(chan error, callers)

	for index := 0; index < callers; index++ {
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			errors <- lookThrough.ensureDirectory(target)
		}()
	}

	waitGroup.Wait()
	close(errors)

	for err := range errors {
		if err != nil {
			t.Fatalf("ensureDirectory() returned error: %v", err)
		}
	}

	info, err := os.Stat(target)
	if err != nil {
		t.Fatalf("failed to stat created directory: %v", err)
	}
	if !info.IsDir() {
		t.Fatalf("created path is not a directory: %s", target)
	}
}

func TestEnsureDirectoryReturnsErrorForInvalidPath(t *testing.T) {
	base := t.TempDir()
	filePath := filepath.Join(base, "file")
	if err := os.WriteFile(filePath, []byte("not a directory"), 0644); err != nil {
		t.Fatalf("failed to create blocking file: %v", err)
	}

	lookThrough := NewLookThrough(variables.LookThroughVars{Workers: 1})
	target := filepath.Join(filePath, "child")

	if err := lookThrough.ensureDirectory(target); err == nil {
		t.Fatal("expected ensureDirectory() to fail when a parent is a file")
	}

	if _, loaded := lookThrough.createdDirs.Load(target); loaded {
		t.Fatal("failed directory creation remained cached")
	}
}
