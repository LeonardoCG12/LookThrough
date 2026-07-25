package lookthrough

import (
	"crypto/sha256"
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/LeonardoCG12/LookThrough/src/utils/handlefile"
	"github.com/LeonardoCG12/LookThrough/src/utils/variables"
)

type testDirectoryEntry struct {
	name string
}

func (e testDirectoryEntry) Name() string               { return e.name }
func (e testDirectoryEntry) IsDir() bool                { return true }
func (e testDirectoryEntry) Type() fs.FileMode          { return fs.ModeDir }
func (e testDirectoryEntry) Info() (fs.FileInfo, error) { return nil, nil }

func TestCountSourceFilesUsesFinalEligibleTotal(t *testing.T) {
	sourceDirectory := t.TempDir()
	destinationDirectory := filepath.Join(sourceDirectory, "new-source")

	files := map[string]string{
		"first.txt":                            "first",
		filepath.Join("nested", "second.txt"):  "second",
		"desktop.ini":                          "ignored",
		filepath.Join("new-source", "old.txt"): "output",
	}

	for name, content := range files {
		path := filepath.Join(sourceDirectory, name)
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatalf("failed to create parent directory: %v", err)
		}
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}
	}

	absoluteDestination, err := filepath.Abs(destinationDirectory)
	if err != nil {
		t.Fatalf("failed to resolve destination path: %v", err)
	}

	lookThrough := NewLookThrough(variables.LookThroughVars{
		MyPath:  sourceDirectory,
		NewPath: destinationDirectory,
	})

	count, err := lookThrough.countSourceFiles(filepath.Clean(absoluteDestination))
	if err != nil {
		t.Fatalf("countSourceFiles() returned error: %v", err)
	}

	if count != 2 {
		t.Fatalf("countSourceFiles() = %d; want 2", count)
	}
}

func TestLoadExistingFilesIgnoresInterruptedSafeCopyFiles(t *testing.T) {
	destinationDirectory := t.TempDir()

	realPath := filepath.Join(destinationDirectory, "real.txt")
	if err := os.WriteFile(realPath, []byte("complete"), 0644); err != nil {
		t.Fatalf("failed to create real destination file: %v", err)
	}

	temporaryPath := filepath.Join(destinationDirectory, ".lookthrough-incomplete-copy-stale")
	if err := os.WriteFile(temporaryPath, []byte("incomplete"), 0644); err != nil {
		t.Fatalf("failed to create interrupted temporary file: %v", err)
	}

	lookThrough := NewLookThrough(variables.LookThroughVars{
		NewPath: destinationDirectory,
		Workers: 1,
	})

	if err := lookThrough.loadExistingFiles(); err != nil {
		t.Fatalf("loadExistingFiles() returned error: %v", err)
	}

	if len(lookThrough.Vars.HashMap) != 1 {
		t.Fatalf("loaded hash count = %d; want 1", len(lookThrough.Vars.HashMap))
	}

	if _, exists := lookThrough.Vars.NameMap["real.txt"]; !exists {
		t.Fatal("expected real destination file to be indexed")
	}

	if _, exists := lookThrough.Vars.NameMap[filepath.Base(temporaryPath)]; exists {
		t.Fatal("did not expect interrupted temporary file to be indexed")
	}
}

func TestFirstErrorPreservesFirstNonNilError(t *testing.T) {
	var first firstError
	first.set(nil)
	first.set(errors.New("first"))
	first.set(errors.New("second"))

	if err := first.get(); err == nil || err.Error() != "first" {
		t.Fatalf("firstError.get() = %v; want first error", err)
	}
}

func TestShouldIgnoreFileIsCaseInsensitive(t *testing.T) {
	for _, name := range []string{"desktop.ini", "DESKTOP.INI", "Thumbs.DB", ".DS_Store"} {
		if !shouldIgnoreFile(name) {
			t.Fatalf("shouldIgnoreFile(%q) = false; want true", name)
		}
	}

	if shouldIgnoreFile("ordinary.txt") {
		t.Fatal("shouldIgnoreFile() ignored an ordinary file")
	}
}

func TestLookForFilesDeduplicatesSortsAndVerifies(t *testing.T) {
	sourceDirectory := t.TempDir()
	destinationDirectory := filepath.Join(sourceDirectory, "new-source")

	files := map[string][]byte{
		filepath.Join("first", "report.txt"):  []byte("alpha"),
		filepath.Join("second", "report.txt"): []byte("beta"),
		"duplicate.txt":                       []byte("alpha"),
		"photo.JPG":                           []byte("image"),
		"desktop.ini":                         []byte("ignored"),
	}

	for name, content := range files {
		path := filepath.Join(sourceDirectory, name)
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatalf("failed to create parent directory: %v", err)
		}
		if err := os.WriteFile(path, content, 0644); err != nil {
			t.Fatalf("failed to create source file: %v", err)
		}
	}

	lookThrough := NewLookThrough(variables.LookThroughVars{
		MyPath:      sourceDirectory,
		NewPath:     destinationDirectory,
		SafeCopy:    true,
		SortFile:    true,
		VerifyFiles: true,
		Workers:     2,
	})

	if err := lookThrough.LookForFiles(false); err != nil {
		t.Fatalf("LookForFiles() returned error: %v", err)
	}

	if lookThrough.Vars.FileCount != 4 {
		t.Fatalf("FileCount = %d; want 4", lookThrough.Vars.FileCount)
	}
	if lookThrough.Vars.HashCount != 3 {
		t.Fatalf("HashCount = %d; want 3", lookThrough.Vars.HashCount)
	}

	wantDigests := map[[32]byte]struct{}{
		sha256.Sum256([]byte("alpha")): {},
		sha256.Sum256([]byte("beta")):  {},
		sha256.Sum256([]byte("image")): {},
	}
	gotDigests := make(map[[32]byte]struct{})
	categoryCounts := make(map[string]int)

	err := filepath.WalkDir(destinationDirectory, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}
		if handlefile.IsTemporaryCopyFile(entry.Name()) {
			t.Fatalf("temporary copy file remained after successful run: %s", path)
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		gotDigests[sha256.Sum256(content)] = struct{}{}

		relative, err := filepath.Rel(destinationDirectory, path)
		if err != nil {
			return err
		}
		categoryCounts[filepath.Dir(relative)]++
		return nil
	})
	if err != nil {
		t.Fatalf("failed to inspect destination: %v", err)
	}

	if !reflect.DeepEqual(gotDigests, wantDigests) {
		t.Fatalf("destination digests = %x; want %x", gotDigests, wantDigests)
	}
	if categoryCounts["Documents"] != 2 || categoryCounts["Images"] != 1 {
		t.Fatalf("category counts = %#v; want Documents=2 and Images=1", categoryCounts)
	}
}

func TestLookForFilesReturnsErrorWhenNoEligibleFilesExist(t *testing.T) {
	sourceDirectory := t.TempDir()
	if err := os.WriteFile(filepath.Join(sourceDirectory, "desktop.ini"), []byte("ignored"), 0644); err != nil {
		t.Fatalf("failed to create ignored file: %v", err)
	}

	lookThrough := NewLookThrough(variables.LookThroughVars{
		MyPath:  sourceDirectory,
		NewPath: filepath.Join(sourceDirectory, "new-source"),
		Workers: 1,
	})

	if err := lookThrough.LookForFiles(false); err == nil {
		t.Fatal("expected LookForFiles() to fail when no eligible files exist")
	}
}

func TestHandleSourceWalkErrorFailsByDefault(t *testing.T) {
	sourcePath := filepath.Join("source")
	lookThrough := NewLookThrough(variables.LookThroughVars{
		MyPath:  sourcePath,
		Workers: 1,
	})

	walkErr := errors.New("permission denied")
	err := lookThrough.handleSourceWalkError(
		filepath.Join(sourcePath, "private"),
		testDirectoryEntry{name: "private"},
		walkErr,
	)

	if err == nil {
		t.Fatal("expected source directory access error to stop the scan")
	}
	if !errors.Is(err, walkErr) {
		t.Fatalf("handleSourceWalkError() error = %v; want wrapped walk error", err)
	}
	if len(lookThrough.skippedDirectoriesSnapshot()) != 0 {
		t.Fatal("did not expect a failed directory to be recorded as skipped")
	}
}

func TestHandleSourceWalkErrorRecordsAndDeduplicatesWhenIgnored(t *testing.T) {
	sourcePath := filepath.Join("source")
	lookThrough := NewLookThrough(variables.LookThroughVars{
		MyPath:                sourcePath,
		IgnoreDirectoryErrors: true,
		Workers:               1,
	})

	path := filepath.Join(sourcePath, "private")
	entry := testDirectoryEntry{name: "private"}

	for _, walkErr := range []error{errors.New("permission denied"), errors.New("permission denied")} {
		if err := lookThrough.handleSourceWalkError(path, entry, walkErr); err != fs.SkipDir {
			t.Fatalf("handleSourceWalkError() = %v; want fs.SkipDir", err)
		}
	}

	directories := lookThrough.skippedDirectoriesSnapshot()
	if len(directories) != 1 {
		t.Fatalf("skipped directory count = %d; want 1", len(directories))
	}
	if directories[0].path != filepath.Clean(path) {
		t.Fatalf("skipped directory path = %q; want %q", directories[0].path, filepath.Clean(path))
	}
	if directories[0].reason != "permission denied" {
		t.Fatalf("skipped directory reason = %q; want permission denied", directories[0].reason)
	}
}

func TestHandleSourceWalkErrorDoesNotIgnoreFileErrors(t *testing.T) {
	sourcePath := filepath.Join("source")
	lookThrough := NewLookThrough(variables.LookThroughVars{
		MyPath:                sourcePath,
		IgnoreDirectoryErrors: true,
		Workers:               1,
	})

	walkErr := errors.New("file disappeared")
	err := lookThrough.handleSourceWalkError(filepath.Join(sourcePath, "file.txt"), nil, walkErr)

	if !errors.Is(err, walkErr) {
		t.Fatalf("handleSourceWalkError() error = %v; want file error", err)
	}
	if len(lookThrough.skippedDirectoriesSnapshot()) != 0 {
		t.Fatal("did not expect a file error to be recorded as a skipped directory")
	}
}

func TestPrintSkippedDirectoriesReportsSortedUniqueEntries(t *testing.T) {
	lookThrough := NewLookThrough(variables.LookThroughVars{Workers: 1})
	alphaPath := filepath.Join("source", "alpha")
	zetaPath := filepath.Join("source", "zeta")
	lookThrough.recordSkippedDirectory(zetaPath, errors.New("permission denied"))
	lookThrough.recordSkippedDirectory(alphaPath, errors.New("not accessible"))
	lookThrough.recordSkippedDirectory(zetaPath, errors.New("permission denied"))

	originalStdout := os.Stdout
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create stdout pipe: %v", err)
	}

	os.Stdout = writer
	lookThrough.printSkippedDirectories()
	if err := writer.Close(); err != nil {
		t.Fatalf("failed to close stdout writer: %v", err)
	}
	os.Stdout = originalStdout

	output, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("failed to read stdout: %v", err)
	}

	text := string(output)
	if !strings.Contains(text, "SKIPPED SOURCE DIRECTORIES: 2") {
		t.Fatalf("report output = %q; want skipped directory count", text)
	}
	alphaIndex := strings.Index(text, alphaPath+": not accessible")
	zetaIndex := strings.Index(text, zetaPath+": permission denied")
	if alphaIndex < 0 || zetaIndex < 0 {
		t.Fatalf("report output = %q; want both skipped directories", text)
	}
	if alphaIndex > zetaIndex {
		t.Fatalf("report output = %q; want paths sorted", text)
	}
}
