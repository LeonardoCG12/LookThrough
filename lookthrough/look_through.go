package lookthrough

import (
	"crypto/md5"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"

	"github.com/LeonardoCG12/LookThrough/utils/gethardware"
	"github.com/LeonardoCG12/LookThrough/utils/getpath"
	"github.com/LeonardoCG12/LookThrough/utils/handlefile"
	"github.com/LeonardoCG12/LookThrough/utils/progressbar"
	"github.com/LeonardoCG12/LookThrough/variables"
)

const (
	_  = iota
	KB = 1 << (iota * 10)
	MB
	GB
	TB
	StatusDuplicate       = 0
	StatusConflictingName = 1
	StatusNewFile         = 2
)

type LookThrough struct {
	Vars      variables.LookThroughVars
	mu        sync.Mutex
	semaphore chan struct{}
}

func NewLookThrough(vars variables.LookThroughVars) *LookThrough {
	if vars.Mem == nil {
		vars.Mem = make(map[string]int)
	}
	if vars.HashMap == nil {
		vars.HashMap = make(map[string]bool)
	}
	if vars.NameMap == nil {
		vars.NameMap = make(map[string]bool)
	}
	vars.HashList = []variables.FileHash{}
	vars.HashListAll = []variables.FileHash{}

	dynamicLimit := gethardware.CalculateDynamicLimit()

	return &LookThrough{
		Vars:      vars,
		semaphore: make(chan struct{}, dynamicLimit),
	}
}

func (l *LookThrough) getMD5Checksum(filePath, fileName string, fileSize int64) error {
	fin, err := handlefile.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("Failed to read file %s: %w", fileName, err)
	}
	defer fin.Close()

	hasher := md5.New()
	if _, err := io.Copy(hasher, fin); err != nil {
		return fmt.Errorf("Failed to hash file %s: %w", fileName, err)
	}
	md5Sum := fmt.Sprintf("%x", hasher.Sum(nil))
	return l.saveHash(fileName, filePath, md5Sum, fileSize)
}

func (l *LookThrough) saveHash(fileName, filePath, md5Sum string, fileSize int64) error {
	l.mu.Lock()

	l.Vars.Num = ""
	status := l.lookForHashes(fileName, md5Sum)
	var newFilePath string

	switch status {
	case StatusConflictingName:
		l.Vars.HashCount++
		l.Vars.HashList = append(l.Vars.HashList, variables.FileHash{Name: fileName, Hash: md5Sum})
		l.Vars.HashMap[md5Sum] = true

		value := l.Vars.Mem[fileName]
		if value > 0 {
			l.Vars.Mem[fileName] = value + 1
			l.Vars.Num = fmt.Sprintf("%d", value+1)
		} else {
			l.Vars.Mem[fileName] = 1
			l.Vars.Num = "1"
		}
		newFilePath = getpath.GetNewFilePath(l.Vars.NewPath, l.Vars.Separator, fileName, l.Vars.Num, 1)
		l.Vars.SizeCount += fileSize

	case StatusNewFile:
		l.Vars.HashCount++
		l.Vars.Mem[fileName] = 0
		l.Vars.HashList = append(l.Vars.HashList, variables.FileHash{Name: fileName, Hash: md5Sum})
		l.Vars.HashMap[md5Sum] = true

		newFilePath = getpath.GetNewFilePath(l.Vars.NewPath, l.Vars.Separator, fileName, "", 2)
		l.Vars.SizeCount += fileSize
	}

	l.Vars.NameMap[fileName] = true

	l.Vars.HashListAll = append(l.Vars.HashListAll, variables.FileHash{Name: fileName, Hash: md5Sum})
	l.Vars.TotalSizeCount += fileSize

	l.mu.Unlock()

	if status == StatusConflictingName || status == StatusNewFile {
		if err := handlefile.CopyFile(filePath, newFilePath); err != nil {
			return fmt.Errorf("Failed to copy file %s: %w", fileName, err)
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

func (l *LookThrough) LookForFiles(showProgress bool) error {
	newPathDir := filepath.Base(l.Vars.NewPath)

	var wg sync.WaitGroup
	var routineErr error
	var errMu sync.Mutex

	var totalFiles int32 = 0
	var processedFiles int32 = 0

	if showProgress {
		filepath.WalkDir(l.Vars.MyPath, func(path string, d fs.DirEntry, err error) error {
			if err != nil || d.IsDir() {
				return nil
			}
			name := d.Name()
			if name == "desktop.ini" || name == "thumbs.db" || name == ".DS_Store" {
				return nil
			}
			parentDir := filepath.Base(filepath.Dir(path))
			if parentDir == newPathDir {
				return nil
			}
			totalFiles++
			return nil
		})

		progressbar.PrintProgressBar(0, totalFiles)
	}

	err := filepath.WalkDir(l.Vars.MyPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("Error accessing path %s: %w", path, err)
		}

		if d.IsDir() {
			return nil
		}
		name := d.Name()
		if name == "desktop.ini" || name == "thumbs.db" || name == ".DS_Store" {
			return nil
		}
		parentDir := filepath.Base(filepath.Dir(path))
		if parentDir == newPathDir {
			return nil
		}

		info, err := d.Info()
		if err != nil {
			return fmt.Errorf("Failed to get info for file %s: %w", name, err)
		}

		l.mu.Lock()
		l.Vars.FileCount++
		l.mu.Unlock()

		fileName := name
		size := info.Size()

		gethardware.ThrottleIfMemoryHigh()

		wg.Add(1)
		go func() {
			defer wg.Done()

			l.semaphore <- struct{}{}
			defer func() { <-l.semaphore }()

			if err := l.getMD5Checksum(path, fileName, size); err != nil {
				errMu.Lock()
				if routineErr == nil {
					routineErr = err
				}
				errMu.Unlock()
			}

			if showProgress {
				current := atomic.AddInt32(&processedFiles, 1)
				progressbar.PrintProgressBar(current, totalFiles)
			}
		}()

		return nil
	})

	if err != nil {
		return err
	}

	wg.Wait()

	if showProgress {
		fmt.Println()
	}

	if routineErr != nil {
		return routineErr
	}

	if l.verifyFiles() {
		size, unit := getSize(l.Vars.TotalSizeCount - l.Vars.SizeCount)
		fmt.Print("\n[+] SUCCESS\n")
		fmt.Print("[+] ALL FILES HAVE BEEN COPIED\n\n")
		fmt.Printf(">>> Old Files: %d\n", l.Vars.FileCount)
		fmt.Printf(">>> New Files: %d\n", l.Vars.HashCount)
		fmt.Printf(">>> Freed Storage: %.1f%s\n\n", size, unit)
	} else if l.Vars.FileCount == 0 {
		fmt.Print("\n[-] FAIL\n")
		fmt.Print("[-] NO FILES FOUND\n\n")
		os.Remove(l.Vars.NewPath)
	} else {
		fmt.Print("\n[-] FAIL\n")
		fmt.Print("[-] SOMETHING WENT WRONG\n\n")
	}
	return nil
}

func (l *LookThrough) verifyFiles() bool {
	hashSet := make(map[string]bool)
	for _, fh := range l.Vars.HashList {
		hashSet[fh.Hash] = true
	}
	for _, all := range l.Vars.HashListAll {
		if !hashSet[all.Hash] {
			return false
		}
	}
	return true
}

func getSize(size int64) (float64, string) {
	if size < KB {
		return float64(size), "B"
	} else if size < MB {
		return float64(size) / KB, "KB"
	} else if size < GB {
		return float64(size) / MB, "MB"
	} else if size < TB {
		return float64(size) / GB, "GB"
	} else {
		return float64(size) / TB, "TB"
	}
}
