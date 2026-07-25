package lookthrough

import (
	"crypto/sha256"
	"fmt"
	"io"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/LeonardoCG12/LookThrough/src/utils/handlefile"
	"github.com/LeonardoCG12/LookThrough/src/utils/sortfile"
	"github.com/LeonardoCG12/LookThrough/src/utils/variables"
)

const (
	StatusDuplicate       = 0
	StatusConflictingName = 1
	StatusNewFile         = 2
)

var bufferPool = sync.Pool{
	New: func() any {
		buffer := make([]byte, 128*1024)
		return &buffer
	},
}

func (l *LookThrough) getSHA256Checksum(fileName, filePath string) (variables.Digest, error) {
	fin, err := handlefile.ReadFile(filePath)
	if err != nil {
		return variables.Digest{}, fmt.Errorf("failed to open file '%s': %w", fileName, err)
	}
	defer fin.Close()

	hasher := sha256.New()

	bufferPointer := bufferPool.Get().(*[]byte)
	defer bufferPool.Put(bufferPointer)

	if _, err := io.CopyBuffer(hasher, fin, *bufferPointer); err != nil {
		return variables.Digest{}, fmt.Errorf("failed to generate sha256 hash for '%s': %w", fileName, err)
	}

	var digest variables.Digest
	copy(digest[:], hasher.Sum(nil))

	return digest, nil
}

func (l *LookThrough) saveHash(fileName, filePath string, fileSize int64) error {
	digest, err := l.getSHA256Checksum(fileName, filePath)
	if err != nil {
		return err
	}

	status, targetPath, nameKey := l.reserveDestination(fileName, digest, fileSize)
	if status == StatusDuplicate {
		return nil
	}

	targetDirectory := filepath.Dir(targetPath)
	if err := l.ensureDirectory(targetDirectory); err != nil {
		l.rollbackReservation(digest, nameKey)
		return err
	}

	if err := handlefile.CopyFile(filePath, targetPath, l.Vars.SafeCopy); err != nil {
		l.rollbackReservation(digest, nameKey)
		return fmt.Errorf("failed to copy file '%s' to destination: %w", fileName, err)
	}

	l.mu.Lock()

	if l.Vars.VerifyFiles {
		l.Vars.HashList = append(l.Vars.HashList, variables.FileHash{
			Name: filepath.Base(targetPath),
			Hash: digest,
		})
	}

	l.Vars.SizeCount += fileSize
	l.Vars.HashCount++

	l.mu.Unlock()

	return nil
}

func (l *LookThrough) reserveDestination(fileName string, digest variables.Digest, fileSize int64) (status int, targetPath, nameKey string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.Vars.TotalSizeCount += fileSize

	if l.Vars.VerifyFiles {
		l.Vars.HashListAll = append(l.Vars.HashListAll, variables.FileHash{
			Name: fileName,
			Hash: digest,
		})
	}

	if _, exists := l.Vars.HashMap[digest]; exists {
		return StatusDuplicate, "", ""
	}

	targetDirectory := l.Vars.NewPath
	if l.Vars.SortFile {
		targetDirectory = filepath.Join(targetDirectory, sortfile.GetCategory(fileName))
	}

	baseKey := destinationNameKey(l.Vars.NewPath, targetDirectory, fileName)
	reservedName := fileName
	status = StatusNewFile

	if _, exists := l.Vars.NameMap[baseKey]; exists {
		status = StatusConflictingName
		counter := l.Vars.Mem[baseKey] + 1

		for {
			candidateName := addNumericSuffix(fileName, counter)
			candidateKey := destinationNameKey(l.Vars.NewPath, targetDirectory, candidateName)

			if _, inUse := l.Vars.NameMap[candidateKey]; !inUse {
				reservedName = candidateName
				nameKey = candidateKey
				l.Vars.Mem[baseKey] = counter
				break
			}

			counter++
		}
	} else {
		nameKey = baseKey
		l.Vars.Mem[baseKey] = 0
	}

	l.Vars.HashMap[digest] = struct{}{}
	l.Vars.NameMap[nameKey] = struct{}{}

	return status, filepath.Join(targetDirectory, reservedName), nameKey
}

func (l *LookThrough) rollbackReservation(digest variables.Digest, nameKey string) {
	l.mu.Lock()
	delete(l.Vars.HashMap, digest)
	delete(l.Vars.NameMap, nameKey)
	l.mu.Unlock()
}

func destinationNameKey(rootDirectory, targetDirectory, fileName string) string {
	relativeDirectory, err := filepath.Rel(rootDirectory, targetDirectory)
	if err != nil || relativeDirectory == "." {
		return fileName
	}

	return filepath.Join(relativeDirectory, fileName)
}

func addNumericSuffix(fileName string, number int) string {
	extension := filepath.Ext(fileName)
	baseName := strings.TrimSuffix(fileName, extension)

	return baseName + "(" + strconv.Itoa(number) + ")" + extension
}
