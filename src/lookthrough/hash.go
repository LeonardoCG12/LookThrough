package lookthrough

import (
	"crypto/md5"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/LeonardoCG12/LookThrough/src/utils/getpath"
	"github.com/LeonardoCG12/LookThrough/src/utils/handlefile"
	"github.com/LeonardoCG12/LookThrough/src/utils/sortfile"
	"github.com/LeonardoCG12/LookThrough/src/utils/variables"
)

const (
	StatusDuplicate       = 0
	StatusConflictingName = 1
	StatusNewFile         = 2
)

func (l *LookThrough) getMD5Checksum(filePath, fileName string, fileSize int64) error {
	fin, err := handlefile.ReadFile(filePath)

	if err != nil {
		return fmt.Errorf("error reading file %s: %w", fileName, err)
	}

	defer fin.Close()

	hasher := md5.New()

	if _, err := io.Copy(hasher, fin); err != nil {
		return fmt.Errorf("error hashing file %s: %w", fileName, err)
	}

	md5Sum := fmt.Sprintf("%x", hasher.Sum(nil))

	return l.saveHash(fileName, filePath, md5Sum, fileSize)
}

func (l *LookThrough) saveHash(fileName, filePath, md5Sum string, fileSize int64) error {
	l.mu.Lock()

	var newFilePath string

	l.Vars.Num = ""
	status := l.lookForHashes(fileName, md5Sum)

	switch status {
	case StatusConflictingName:
		value := l.Vars.Mem[fileName]

		l.Vars.HashList = append(l.Vars.HashList, variables.FileHash{Name: fileName, Hash: md5Sum})
		l.Vars.HashMap[md5Sum] = true

		l.Vars.HashCount++

		if value > 0 {
			l.Vars.Mem[fileName] = value + 1
			l.Vars.Num = fmt.Sprintf("%d", value+1)
		} else {
			l.Vars.Mem[fileName] = 1
			l.Vars.Num = "1"
		}

		newFilePath = getpath.GetNewFilePath(l.Vars.NewPath, fileName, l.Vars.Num, 1)
		l.Vars.SizeCount += fileSize

	case StatusNewFile:
		l.Vars.Mem[fileName] = 0
		l.Vars.HashList = append(l.Vars.HashList, variables.FileHash{Name: fileName, Hash: md5Sum})
		l.Vars.HashMap[md5Sum] = true
		newFilePath = getpath.GetNewFilePath(l.Vars.NewPath, fileName, "", 2)

		l.Vars.SizeCount += fileSize
		l.Vars.HashCount++
	}

	l.Vars.NameMap[fileName] = true
	l.Vars.HashListAll = append(l.Vars.HashListAll, variables.FileHash{Name: fileName, Hash: md5Sum})

	l.Vars.TotalSizeCount += fileSize

	l.mu.Unlock()

	if status == StatusConflictingName || status == StatusNewFile {

		if l.Vars.SortFile {
			category := sortfile.GetCategory(fileName)
			targetDir := filepath.Join(l.Vars.NewPath, category)

			if err := os.MkdirAll(targetDir, 0755); err != nil {
				return fmt.Errorf("error creating directory %s: %w", targetDir, err)
			}

			newFilePath = filepath.Join(targetDir, filepath.Base(newFilePath))
		}

		if err := handlefile.CopyFile(filePath, newFilePath); err != nil {
			return fmt.Errorf("error copying file %s: %w", fileName, err)
		}
	}

	return nil
}

func (l *LookThrough) lookForHashes(fileName, md5Sum string) int {
	if l.Vars.HashMap[md5Sum] {
		return StatusDuplicate
	}

	if l.Vars.NameMap[fileName] {
		return StatusConflictingName
	}

	return StatusNewFile
}
