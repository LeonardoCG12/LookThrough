package lookthrough

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"

	"github.com/LeonardoCG12/LookThrough/src/utils/gethardware"
	"github.com/LeonardoCG12/LookThrough/src/utils/getsize"
	"github.com/LeonardoCG12/LookThrough/src/utils/progressbar"
	"github.com/LeonardoCG12/LookThrough/src/utils/variables"
)

type fileTask struct {
	path string
	name string
	size int64
}

func (l *LookThrough) indexDestinationFiles() error {
	return filepath.WalkDir(l.Vars.NewPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}

		md5Sum, err := l.getMD5Checksum(d.Name(), path)

		if err != nil {
			return fmt.Errorf("error getting MD5 checksum for file %s: %w", d.Name(), err)
		}

		l.Vars.HashMap[md5Sum] = true
		l.Vars.NameMap[d.Name()] = true
		l.Vars.HashList = append(l.Vars.HashList, variables.FileHash{Name: d.Name(), Hash: md5Sum})

		return nil
	})
}

func (l *LookThrough) LookForFiles(folderInspect bool) error {
	if folderInspect {
		if err := l.indexDestinationFiles(); err != nil {
			return fmt.Errorf("error indexing destination files: %w", err)
		}
	}

	newPathDir := filepath.Base(l.Vars.NewPath)

	var wg sync.WaitGroup
	var routineErr error
	var errMu sync.Mutex

	var processedFiles int32 = 0
	var tasks []fileTask

	err := filepath.WalkDir(l.Vars.MyPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			if d != nil && d.IsDir() {
				return fs.SkipDir
			}

			return nil
		}

		if d.IsDir() {
			if d.Name() == newPathDir {
				return fs.SkipDir
			}

			return nil
		}

		name := d.Name()

		if name == "desktop.ini" || name == "thumbs.db" || name == ".DS_Store" {
			return nil
		}

		info, err := d.Info()

		if err != nil {
			return fmt.Errorf("error getting info for file %s: %w", name, err)
		}

		tasks = append(tasks, fileTask{
			path: path,
			name: name,
			size: info.Size(),
		})

		return nil
	})

	if err != nil {
		return err
	}

	totalFiles := int32(len(tasks))

	l.mu.Lock()

	l.Vars.FileCount = int(totalFiles)

	l.mu.Unlock()

	if l.Vars.ShowProgressBar {
		progressbar.PrintProgressBar(0, totalFiles)
	}

	for _, task := range tasks {
		gethardware.ThrottleIfMemoryHigh()

		l.semaphore <- struct{}{}
		wg.Add(1)

		go func(t fileTask) {
			defer wg.Done()
			defer func() { <-l.semaphore }()

			if err := l.saveHash(t.name, t.path, t.size); err != nil {
				errMu.Lock()

				if routineErr == nil {
					routineErr = err
				}

				errMu.Unlock()
			}

			if l.Vars.ShowProgressBar {
				current := atomic.AddInt32(&processedFiles, 1)

				progressbar.PrintProgressBar(current, totalFiles)
			}
		}(task)
	}

	wg.Wait()

	if l.Vars.ShowProgressBar {
		fmt.Println()
	}

	if routineErr != nil {
		return routineErr
	}

	if l.verifyFiles() && l.Vars.FileCount > 0 {
		size, unit := getsize.GetSize(l.Vars.TotalSizeCount - l.Vars.SizeCount)

		fmt.Print("\n[+] SUCCESS\n[+] ALL FILES HAVE BEEN COPIED\n\n")
		fmt.Printf(">>> Old Files: %d\n", l.Vars.FileCount)
		fmt.Printf(">>> New Files: %d\n", l.Vars.HashCount)
		fmt.Printf(">>> Freed Storage: %.1f%s\n\n", size, unit)

		return nil
	}

	if !l.verifyFiles() {
		fmt.Print("\n[-] FAIL\n[-] FILES ARE MISSING OR HAVE BEEN MODIFIED\n\n")
	} else if l.Vars.FileCount == 0 {
		fmt.Print("\n[-] FAIL\n[-] NO FILES FOUND\n\n")
	} else {
		fmt.Print("\n[-] FAIL\n[-] SOMETHING WENT WRONG\n\n")
	}

	os.RemoveAll(l.Vars.NewPath)

	return fmt.Errorf("unexpected error during file processing")
}
