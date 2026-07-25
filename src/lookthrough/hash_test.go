package lookthrough

import (
	"crypto/sha256"
	"os"
	"path/filepath"
	"testing"

	"github.com/LeonardoCG12/LookThrough/src/utils/variables"
)

func TestSkippedDuplicateDoesNotBlockLaterUniqueFile(t *testing.T) {
	base := t.TempDir()
	srcDir := filepath.Join(base, "src")
	destDir := filepath.Join(base, "dest")

	if err := os.MkdirAll(srcDir, 0755); err != nil {
		t.Fatalf("failed to create src dir: %v", err)
	}
	if err := os.MkdirAll(destDir, 0755); err != nil {
		t.Fatalf("failed to create dest dir: %v", err)
	}

	original := filepath.Join(srcDir, "original.txt")
	dupContent := filepath.Join(srcDir, "dup-content.txt")
	newContent := filepath.Join(srcDir, "new-content.txt")

	if err := os.WriteFile(original, []byte("HELLO"), 0644); err != nil {
		t.Fatalf("failed to write original: %v", err)
	}
	if err := os.WriteFile(dupContent, []byte("HELLO"), 0644); err != nil {
		t.Fatalf("failed to write dup-content: %v", err)
	}
	if err := os.WriteFile(newContent, []byte("WORLD"), 0644); err != nil {
		t.Fatalf("failed to write new-content: %v", err)
	}

	l := NewLookThrough(variables.LookThroughVars{
		NewPath: destDir,
	})

	if err := l.saveHash("a.txt", original, 5); err != nil {
		t.Fatalf("saveHash(a.txt) failed: %v", err)
	}

	if err := l.saveHash("b.txt", dupContent, 5); err != nil {
		t.Fatalf("saveHash(b.txt dup) failed: %v", err)
	}

	if err := l.saveHash("b.txt", newContent, 5); err != nil {
		t.Fatalf("saveHash(b.txt new) failed: %v", err)
	}

	if _, err := os.Stat(filepath.Join(destDir, "b.txt")); err != nil {
		t.Errorf("expected %s to exist, got error: %v", filepath.Join(destDir, "b.txt"), err)
	}

	if _, err := os.Stat(filepath.Join(destDir, "b(1).txt")); err == nil {
		t.Errorf("did not expect %s to exist: a skipped duplicate must not consume the name \"b.txt\"", filepath.Join(destDir, "b(1).txt"))
	}
}

func TestGetSHA256Checksum(t *testing.T) {
	directory := t.TempDir()
	path := filepath.Join(directory, "file.txt")
	content := []byte("lookthrough")

	if err := os.WriteFile(path, content, 0644); err != nil {
		t.Fatalf("failed to create source file: %v", err)
	}

	lookThrough := NewLookThrough(variables.LookThroughVars{Workers: 1})
	got, err := lookThrough.getSHA256Checksum("file.txt", path)
	if err != nil {
		t.Fatalf("getSHA256Checksum() returned error: %v", err)
	}

	want := sha256.Sum256(content)
	if got != want {
		t.Fatalf("getSHA256Checksum() = %x; want %x", got, want)
	}
}

func TestReserveDestinationFindsFirstAvailableSuffix(t *testing.T) {
	destination := t.TempDir()
	lookThrough := NewLookThrough(variables.LookThroughVars{
		NewPath: destination,
		Workers: 1,
		NameMap: map[string]struct{}{
			"report.txt":    {},
			"report(1).txt": {},
		},
		Mem: map[string]int{"report.txt": 0},
	})
	digest := variables.Digest{7}

	status, targetPath, nameKey := lookThrough.reserveDestination("report.txt", digest, 10)

	if status != StatusConflictingName {
		t.Fatalf("status = %d; want %d", status, StatusConflictingName)
	}
	if want := filepath.Join(destination, "report(2).txt"); targetPath != want {
		t.Fatalf("targetPath = %q; want %q", targetPath, want)
	}
	if nameKey != "report(2).txt" {
		t.Fatalf("nameKey = %q; want %q", nameKey, "report(2).txt")
	}
	if _, exists := lookThrough.Vars.HashMap[digest]; !exists {
		t.Fatal("reserved digest was not recorded")
	}
	if _, exists := lookThrough.Vars.NameMap[nameKey]; !exists {
		t.Fatal("reserved destination name was not recorded")
	}
}

func TestRollbackReservationRemovesHashAndName(t *testing.T) {
	digest := variables.Digest{9}
	lookThrough := NewLookThrough(variables.LookThroughVars{
		Workers: 1,
		HashMap: map[variables.Digest]struct{}{digest: {}},
		NameMap: map[string]struct{}{"file.txt": {}},
	})

	lookThrough.rollbackReservation(digest, "file.txt")

	if _, exists := lookThrough.Vars.HashMap[digest]; exists {
		t.Fatal("rollbackReservation() did not remove digest")
	}
	if _, exists := lookThrough.Vars.NameMap["file.txt"]; exists {
		t.Fatal("rollbackReservation() did not remove destination name")
	}
}

func TestDestinationNameKey(t *testing.T) {
	root := filepath.Join("root", "output")

	if got := destinationNameKey(root, root, "file.txt"); got != "file.txt" {
		t.Fatalf("destinationNameKey() at root = %q; want %q", got, "file.txt")
	}

	target := filepath.Join(root, "Documents")
	want := filepath.Join("Documents", "file.txt")
	if got := destinationNameKey(root, target, "file.txt"); got != want {
		t.Fatalf("destinationNameKey() in category = %q; want %q", got, want)
	}
}

func TestAddNumericSuffix(t *testing.T) {
	tests := []struct {
		fileName string
		number   int
		want     string
	}{
		{fileName: "report.txt", number: 1, want: "report(1).txt"},
		{fileName: "archive.tar.gz", number: 2, want: "archive.tar(2).gz"},
		{fileName: "README", number: 3, want: "README(3)"},
	}

	for _, test := range tests {
		if got := addNumericSuffix(test.fileName, test.number); got != test.want {
			t.Fatalf("addNumericSuffix(%q, %d) = %q; want %q", test.fileName, test.number, got, test.want)
		}
	}
}
