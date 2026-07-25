package progressbar

import (
	"io"
	"os"
	"strings"
	"testing"
)

func TestPrintProgressBarRendersSampledValue(t *testing.T) {
	originalStdout := os.Stdout
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}

	os.Stdout = writer
	PrintProgressBar(37, 1000)

	if err := writer.Close(); err != nil {
		t.Fatalf("failed to close writer: %v", err)
	}
	os.Stdout = originalStdout

	output, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("failed to read progress output: %v", err)
	}

	if !strings.Contains(string(output), "3.7% (37/1000)") {
		t.Fatalf("progress output = %q; want sampled value", output)
	}
}

func TestPrintProgressBarClampsCurrentValue(t *testing.T) {
	negative := captureProgressOutput(t, -5, 100)
	if !strings.Contains(negative, "0.0% (0/100)") {
		t.Fatalf("negative progress output = %q; want clamped zero", negative)
	}

	overflow := captureProgressOutput(t, 150, 100)
	if !strings.Contains(overflow, "100.0% (100/100)") {
		t.Fatalf("overflow progress output = %q; want clamped total", overflow)
	}
}

func TestPrintProgressBarDoesNotRenderWithoutTotal(t *testing.T) {
	if output := captureProgressOutput(t, 1, 0); output != "" {
		t.Fatalf("progress output = %q; want empty output", output)
	}
}

func captureProgressOutput(t *testing.T, current, total int64) string {
	t.Helper()

	originalStdout := os.Stdout
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}
	os.Stdout = writer
	defer func() {
		os.Stdout = originalStdout
		_ = reader.Close()
	}()

	PrintProgressBar(current, total)

	if err := writer.Close(); err != nil {
		t.Fatalf("failed to close writer: %v", err)
	}

	output, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("failed to read progress output: %v", err)
	}
	return string(output)
}
