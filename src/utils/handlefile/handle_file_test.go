package handlefile

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestCopyFileCopiesContent(t *testing.T) {
	tests := []struct {
		name     string
		safeCopy bool
	}{
		{name: "direct", safeCopy: false},
		{name: "safe", safeCopy: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			directory := t.TempDir()
			sourcePath := filepath.Join(directory, "source.txt")
			destinationPath := filepath.Join(directory, "destination.txt")
			content := []byte("lookthrough")

			if err := os.WriteFile(sourcePath, content, 0640); err != nil {
				t.Fatalf("failed to create source file: %v", err)
			}

			if err := CopyFile(sourcePath, destinationPath, test.safeCopy); err != nil {
				t.Fatalf("CopyFile() returned error: %v", err)
			}

			copiedContent, err := os.ReadFile(destinationPath)
			if err != nil {
				t.Fatalf("failed to read destination file: %v", err)
			}

			if string(copiedContent) != string(content) {
				t.Fatalf("destination content = %q; want %q", copiedContent, content)
			}

			entries, err := os.ReadDir(directory)
			if err != nil {
				t.Fatalf("failed to read test directory: %v", err)
			}

			for _, entry := range entries {
				if IsTemporaryCopyFile(entry.Name()) {
					t.Fatalf("temporary copy file was not removed: %s", entry.Name())
				}
			}
		})
	}
}

func TestIsTemporaryCopyFile(t *testing.T) {
	if !IsTemporaryCopyFile(temporaryCopyPrefix + "123") {
		t.Fatal("expected internal temporary copy name to be detected")
	}

	if IsTemporaryCopyFile("ordinary-file.txt") {
		t.Fatal("did not expect ordinary file name to be detected as temporary")
	}
}

func TestCopyFileReturnsErrorForMissingSource(t *testing.T) {
	for _, safeCopy := range []bool{false, true} {
		t.Run(map[bool]string{false: "direct", true: "safe"}[safeCopy], func(t *testing.T) {
			destination := filepath.Join(t.TempDir(), "destination.txt")
			if err := CopyFile(filepath.Join(t.TempDir(), "missing.txt"), destination, safeCopy); err == nil {
				t.Fatal("expected CopyFile() to fail for a missing source")
			}
			if _, err := os.Stat(destination); !os.IsNotExist(err) {
				t.Fatalf("destination exists after failed copy; stat error = %v", err)
			}
		})
	}
}

func TestSafeCopyPreservesSourcePermissions(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX permission bits are not reliably represented on Windows")
	}

	directory := t.TempDir()
	sourcePath := filepath.Join(directory, "source.txt")
	destinationPath := filepath.Join(directory, "destination.txt")

	if err := os.WriteFile(sourcePath, []byte("content"), 0640); err != nil {
		t.Fatalf("failed to create source file: %v", err)
	}
	if err := CopyFile(sourcePath, destinationPath, true); err != nil {
		t.Fatalf("CopyFile() returned error: %v", err)
	}

	info, err := os.Stat(destinationPath)
	if err != nil {
		t.Fatalf("failed to stat destination: %v", err)
	}
	if got := info.Mode().Perm(); got != 0640 {
		t.Fatalf("destination permissions = %04o; want 0640", got)
	}
}
