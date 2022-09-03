package handlefile

import (
	"io"
	"log"
	"os"
)

func ReadFile(filePath string) *os.File {
	fin, err := os.Open(filePath)

	if err != nil {
		log.Fatal(err)
	}

	return fin
}

func CopyFile(filePath, newFilePath string) {
	fin, err := os.Open(filePath)

	if err != nil {
		log.Fatal(err)
	}

	defer fin.Close()

	fout, err := os.Create(newFilePath)

	if err != nil {
		log.Fatal(err)
	}

	defer fout.Close()

	_, err = io.Copy(fout, fin)

	if err != nil {
		log.Fatal(err)
	}
}
