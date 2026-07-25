package handlefolder

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMakeNewDirCreatesMissingDirectory(t *testing.T) {
	path := filepath.Join(t.TempDir(), "new")

	inspect, err := MakeNewDir(path)
	if err != nil {
		t.Fatalf("MakeNewDir() returned error: %v", err)
	}
	if inspect {
		t.Fatal("MakeNewDir() requested inspection for a newly created directory")
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("failed to stat created directory: %v", err)
	}
	if !info.IsDir() {
		t.Fatalf("created path is not a directory: %s", path)
	}
}

func TestMakeNewDirAmendsExistingDirectory(t *testing.T) {
	path := t.TempDir()
	existingFile := filepath.Join(path, "existing.txt")
	if err := os.WriteFile(existingFile, []byte("keep"), 0644); err != nil {
		t.Fatalf("failed to create existing file: %v", err)
	}

	withStdin(t, "a\n", func() {
		inspect, err := MakeNewDir(path)
		if err != nil {
			t.Fatalf("MakeNewDir() returned error: %v", err)
		}
		if !inspect {
			t.Fatal("MakeNewDir() did not request inspection for amend mode")
		}
	})

	if _, err := os.Stat(existingFile); err != nil {
		t.Fatalf("amend mode removed existing file: %v", err)
	}
}

func TestMakeNewDirDeletesExistingContents(t *testing.T) {
	path := t.TempDir()
	existingFile := filepath.Join(path, "existing.txt")
	if err := os.WriteFile(existingFile, []byte("remove"), 0644); err != nil {
		t.Fatalf("failed to create existing file: %v", err)
	}

	withStdin(t, "invalid\nd\n", func() {
		inspect, err := MakeNewDir(path)
		if err != nil {
			t.Fatalf("MakeNewDir() returned error: %v", err)
		}
		if inspect {
			t.Fatal("MakeNewDir() requested inspection for delete mode")
		}
	})

	if _, err := os.Stat(existingFile); !os.IsNotExist(err) {
		t.Fatalf("delete mode did not remove existing file; stat error = %v", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("failed to stat recreated directory: %v", err)
	}
	if !info.IsDir() {
		t.Fatalf("recreated path is not a directory: %s", path)
	}
}

func TestMakeNewDirFailsWhenPathIsAFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "file")
	if err := os.WriteFile(path, []byte("file"), 0644); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}

	if _, err := MakeNewDir(path); err == nil {
		t.Fatal("expected MakeNewDir() to fail when the target is a file")
	}
}

func withStdin(t *testing.T, input string, run func()) {
	t.Helper()

	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create stdin pipe: %v", err)
	}

	originalStdin := os.Stdin
	os.Stdin = reader
	defer func() {
		os.Stdin = originalStdin
		_ = reader.Close()
	}()

	if _, err := writer.WriteString(input); err != nil {
		t.Fatalf("failed to write stdin input: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("failed to close stdin writer: %v", err)
	}

	run()
}
