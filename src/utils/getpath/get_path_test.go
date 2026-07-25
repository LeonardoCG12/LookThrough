package getpath

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetPathUsesFlagBeforePositionalArgument(t *testing.T) {
	flagPath := filepath.Join(t.TempDir(), "flag")
	positionalPath := filepath.Join(t.TempDir(), "positional")

	got, err := GetPath(flagPath, []string{positionalPath})
	if err != nil {
		t.Fatalf("GetPath() returned error: %v", err)
	}

	want, err := filepath.Abs(flagPath)
	if err != nil {
		t.Fatalf("failed to resolve expected path: %v", err)
	}

	if got != want {
		t.Fatalf("GetPath() = %q; want %q", got, want)
	}
}

func TestGetPathUsesPositionalArgument(t *testing.T) {
	positionalPath := filepath.Join(t.TempDir(), "positional")

	got, err := GetPath("", []string{positionalPath})
	if err != nil {
		t.Fatalf("GetPath() returned error: %v", err)
	}

	want, err := filepath.Abs(positionalPath)
	if err != nil {
		t.Fatalf("failed to resolve expected path: %v", err)
	}

	if got != want {
		t.Fatalf("GetPath() = %q; want %q", got, want)
	}
}

func TestGetPathReadsInteractiveInput(t *testing.T) {
	inputPath := filepath.Join(t.TempDir(), "interactive")
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create input pipe: %v", err)
	}

	originalStdin := os.Stdin
	os.Stdin = reader
	t.Cleanup(func() {
		os.Stdin = originalStdin
		_ = reader.Close()
	})

	if _, err := writer.WriteString("\n" + inputPath + "\n"); err != nil {
		t.Fatalf("failed to write interactive input: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("failed to close input writer: %v", err)
	}

	got, err := GetPath("", nil)
	if err != nil {
		t.Fatalf("GetPath() returned error: %v", err)
	}

	want, err := filepath.Abs(inputPath)
	if err != nil {
		t.Fatalf("failed to resolve expected path: %v", err)
	}

	if got != want {
		t.Fatalf("GetPath() = %q; want %q", got, want)
	}
}

func TestGetNewPath(t *testing.T) {
	target := filepath.Join(t.TempDir(), "MyFolder")
	want := filepath.Join(target, "new-MyFolder")

	if got := GetNewPath(target); got != want {
		t.Fatalf("GetNewPath() = %q; want %q", got, want)
	}
}
