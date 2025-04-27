package lookthrough

import (
	"crypto/md5"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"

	"io/fs"
	"slices"

	"golang.org/x/sync/errgroup"

	"github.com/LeonardoCG12/LookThrough/utils/getpath"
	"github.com/LeonardoCG12/LookThrough/utils/handlefile"
	"github.com/LeonardoCG12/LookThrough/variables"
)

const (
	_  = iota
	KB = 1 << (iota * 10)
	MB
	GB
	TB
)

type LookThrough struct {
	Vars variables.LookThroughVars
	mu   sync.Mutex
}

func NewLookThrough(vars variables.LookThroughVars) *LookThrough {
	if vars.Mem == nil {
		vars.Mem = make(map[string]int)
	}
	vars.HashList = []variables.FileHash{}
	vars.HashListAll = []variables.FileHash{}
	return &LookThrough{Vars: vars}
}

func (l *LookThrough) getMD5Checksum(filePath, fileName string, fileSize int64) error {
	fin, err := handlefile.ReadFile(filePath)
	if err != nil {
		return err
	}
	defer fin.Close()

	hasher := md5.New()
	if _, err := io.Copy(hasher, fin); err != nil {
		return err
	}
	md5Sum := fmt.Sprintf("%x", hasher.Sum(nil))
	return l.saveHash(fileName, filePath, md5Sum, fileSize)
}

func (l *LookThrough) saveHash(fileName, filePath, md5Sum string, fileSize int64) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.Vars.Num = ""
	status := l.lookForHashes(fileName, md5Sum)
	var newFilePath string

	if status == 1 {
		l.Vars.HashCount++
		l.Vars.HashList = append(l.Vars.HashList, variables.FileHash{Name: fileName, Hash: md5Sum})
		value := l.Vars.Mem[fileName]
		if value > 0 {
			l.Vars.Mem[fileName] = value + 1
			l.Vars.Num = fmt.Sprintf("%d", value+1)
		} else {
			l.Vars.Mem[fileName] = 1
			l.Vars.Num = ""
		}
		newFilePath = getpath.GetNewFilePath(l.Vars.NewPath, l.Vars.Separator, fileName, l.Vars.Num, status)
		l.Vars.SizeCount += fileSize
	} else if status == 2 {
		l.Vars.HashCount++
		l.Vars.Mem[fileName] = 0
		l.Vars.HashList = append(l.Vars.HashList, variables.FileHash{Name: fileName, Hash: md5Sum})
		newFilePath = getpath.GetNewFilePath(l.Vars.NewPath, l.Vars.Separator, fileName, "", status)
		l.Vars.SizeCount += fileSize
	}

	l.Vars.HashListAll = append(l.Vars.HashListAll, variables.FileHash{Name: fileName, Hash: md5Sum})
	l.Vars.TotalSizeCount += fileSize

	if status == 1 || status == 2 {
		if err := handlefile.CopyFile(filePath, newFilePath); err != nil {
			return err
		}
	}

	return nil
}

func (l *LookThrough) lookForHashes(fileName, md5Sum string) int {
	for _, fh := range l.Vars.HashList {
		if fh.Hash == md5Sum {
			return 0
		}
	}
	for _, fh := range l.Vars.HashList {
		if fh.Name == fileName {
			return 1
		}
	}
	return 2
}

func (l *LookThrough) LookForFiles() error {
	newPathDir := filepath.Base(l.Vars.NewPath)

	var g errgroup.Group
	err := filepath.WalkDir(l.Vars.MyPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
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
			return err
		}

		l.mu.Lock()
		l.Vars.FileCount++
		l.mu.Unlock()

		fileName := name
		size := info.Size()
		g.Go(func() error {
			return l.getMD5Checksum(path, fileName, size)
		})

		return nil
	})
	if err != nil {
		return err
	}

	if err := g.Wait(); err != nil {
		return err
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
	for _, all := range l.Vars.HashListAll {
		hashes := make([]string, len(l.Vars.HashList))
		for i, fh := range l.Vars.HashList {
			hashes[i] = fh.Hash
		}
		if !slices.Contains(hashes, all.Hash) {
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
