package handlefile

import (
	"io"
	"os"
)

func ReadFile(filePath string) (*os.File, error) {
	return os.Open(filePath)
}

func CopyFile(filePath, newFilePath string) error {
	fin, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer fin.Close()

	fout, err := os.Create(newFilePath)
	if err != nil {
		return err
	}
	defer fout.Close()

	_, err = io.Copy(fout, fin)
	return err
}
