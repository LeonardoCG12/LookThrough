package handlefile

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const temporaryCopyPrefix = ".lookthrough-incomplete-copy-"

func ReadFile(filePath string) (*os.File, error) {
	return os.Open(filePath)
}

func CopyFile(filePath, newFilePath string, safeCopy bool) error {
	if safeCopy {
		return copyFileSafely(filePath, newFilePath)
	}

	return copyFileDirectly(filePath, newFilePath)
}

func IsTemporaryCopyFile(fileName string) bool {
	return strings.HasPrefix(fileName, temporaryCopyPrefix)
}

func copyFileDirectly(filePath, newFilePath string) error {
	source, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open source file for copying: %w", err)
	}
	defer source.Close()

	destination, err := os.Create(newFilePath)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}

	_, copyErr := io.Copy(destination, source)
	closeErr := destination.Close()

	if copyErr != nil {
		_ = os.Remove(newFilePath)
		return fmt.Errorf("failed to write data to destination file: %w", copyErr)
	}

	if closeErr != nil {
		_ = os.Remove(newFilePath)
		return fmt.Errorf("failed to close destination file: %w", closeErr)
	}

	return nil
}

func copyFileSafely(filePath, newFilePath string) error {
	source, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open source file for safe copying: %w", err)
	}
	defer source.Close()

	sourceInfo, err := source.Stat()
	if err != nil {
		return fmt.Errorf("failed to read source file metadata for safe copying: %w", err)
	}

	temporary, err := os.CreateTemp(filepath.Dir(newFilePath), temporaryCopyPrefix+"*")
	if err != nil {
		return fmt.Errorf("failed to create temporary destination file: %w", err)
	}

	temporaryPath := temporary.Name()
	committed := false

	defer func() {
		_ = temporary.Close()

		if !committed {
			_ = os.Remove(temporaryPath)
		}
	}()

	if err := temporary.Chmod(sourceInfo.Mode().Perm()); err != nil {
		return fmt.Errorf("failed to set temporary destination permissions: %w", err)
	}

	if _, err := io.Copy(temporary, source); err != nil {
		return fmt.Errorf("failed to write data to temporary destination file: %w", err)
	}

	if err := temporary.Close(); err != nil {
		return fmt.Errorf("failed to close temporary destination file: %w", err)
	}

	if err := os.Rename(temporaryPath, newFilePath); err != nil {
		return fmt.Errorf("failed to finalize destination file: %w", err)
	}

	committed = true
	return nil
}
