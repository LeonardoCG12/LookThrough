package lookthrough

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/LeonardoCG12/LookThrough/src/utils/getsize"
	"github.com/LeonardoCG12/LookThrough/src/utils/handlefile"
	"github.com/LeonardoCG12/LookThrough/src/utils/progressbar"
	"github.com/LeonardoCG12/LookThrough/src/utils/variables"
)

type fileJob struct {
	path string
	name string
}

type existingFileJob struct {
	path    string
	name    string
	nameKey string
}

type firstError struct {
	mu  sync.Mutex
	err error
}

func (f *firstError) set(err error) {
	if err == nil {
		return
	}

	f.mu.Lock()
	if f.err == nil {
		f.err = err
	}
	f.mu.Unlock()
}

func (f *firstError) get() error {
	f.mu.Lock()
	defer f.mu.Unlock()

	return f.err
}

type skippedDirectory struct {
	path   string
	reason string
}

type progressState struct {
	processed atomic.Int64
	total     int64
}

func (p *progressState) addProcessed() {
	if p != nil {
		p.processed.Add(1)
	}
}

func (p *progressState) snapshot() (processed, total int64) {
	return p.processed.Load(), p.total
}

func (l *LookThrough) loadExistingFiles() error {
	jobs := make(chan existingFileJob, l.workers*2)

	var workers sync.WaitGroup
	var workerError firstError

	for worker := 0; worker < l.workers; worker++ {
		workers.Add(1)

		go func() {
			defer workers.Done()

			for job := range jobs {
				digest, err := l.getSHA256Checksum(job.name, job.path)
				if err != nil {
					workerError.set(fmt.Errorf("failed to hash existing file '%s': %w", job.name, err))
					continue
				}

				l.mu.Lock()
				l.Vars.HashMap[digest] = struct{}{}
				l.Vars.NameMap[job.nameKey] = struct{}{}

				if l.Vars.VerifyFiles {
					l.Vars.HashList = append(l.Vars.HashList, variables.FileHash{
						Name: job.name,
						Hash: digest,
					})
				}

				l.mu.Unlock()
			}
		}()
	}

	walkError := filepath.WalkDir(l.Vars.NewPath, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}

		if handlefile.IsTemporaryCopyFile(entry.Name()) {
			return nil
		}

		nameKey, err := filepath.Rel(l.Vars.NewPath, path)
		if err != nil {
			return fmt.Errorf("failed to determine destination-relative path for '%s': %w", path, err)
		}

		jobs <- existingFileJob{
			path:    path,
			name:    entry.Name(),
			nameKey: nameKey,
		}

		return nil
	})

	close(jobs)
	workers.Wait()

	if walkError != nil {
		return fmt.Errorf("failed to walk destination directory '%s': %w", l.Vars.NewPath, walkError)
	}
	if err := workerError.get(); err != nil {
		return err
	}

	return nil
}

func (l *LookThrough) walkSourceFiles(absoluteNewPath string, visit func(path string, entry fs.DirEntry) error) error {
	return filepath.WalkDir(l.Vars.MyPath, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return l.handleSourceWalkError(path, entry, walkErr)
		}

		cleanPath := filepath.Clean(path)
		if cleanPath == absoluteNewPath || strings.HasPrefix(cleanPath, absoluteNewPath+string(os.PathSeparator)) {
			if entry.IsDir() {
				return fs.SkipDir
			}
			return nil
		}

		if entry.IsDir() || shouldIgnoreFile(entry.Name()) {
			return nil
		}

		return visit(path, entry)
	})
}

func (l *LookThrough) handleSourceWalkError(path string, entry fs.DirEntry, walkErr error) error {
	isDirectory := entry != nil && entry.IsDir()
	if entry == nil && filepath.Clean(path) == filepath.Clean(l.Vars.MyPath) {
		isDirectory = true
	}

	if !isDirectory {
		return walkErr
	}

	if !l.Vars.IgnoreDirectoryErrors {
		return fmt.Errorf("failed to access source directory '%s': %w", path, walkErr)
	}

	l.recordSkippedDirectory(path, walkErr)
	return fs.SkipDir
}

func (l *LookThrough) recordSkippedDirectory(path string, err error) {
	l.scanMu.Lock()
	l.skippedDirectories[filepath.Clean(path)] = err.Error()
	l.scanMu.Unlock()
}

func (l *LookThrough) resetSkippedDirectories() {
	l.scanMu.Lock()
	l.skippedDirectories = make(map[string]string)
	l.scanMu.Unlock()
}

func (l *LookThrough) skippedDirectoriesSnapshot() []skippedDirectory {
	l.scanMu.Lock()
	directories := make([]skippedDirectory, 0, len(l.skippedDirectories))
	for path, reason := range l.skippedDirectories {
		directories = append(directories, skippedDirectory{path: path, reason: reason})
	}
	l.scanMu.Unlock()

	sort.Slice(directories, func(i, j int) bool {
		return directories[i].path < directories[j].path
	})

	return directories
}

func (l *LookThrough) printSkippedDirectories() {
	directories := l.skippedDirectoriesSnapshot()
	if len(directories) == 0 {
		return
	}

	fmt.Printf("\n[!] SKIPPED SOURCE DIRECTORIES: %d\n", len(directories))
	for _, directory := range directories {
		fmt.Printf("[!] %s: %s\n", directory.path, directory.reason)
	}
	fmt.Println()
}

func (l *LookThrough) countSourceFiles(absoluteNewPath string) (int64, error) {
	var fileCount int64

	err := l.walkSourceFiles(absoluteNewPath, func(_ string, _ fs.DirEntry) error {
		fileCount++
		return nil
	})
	if err != nil {
		return 0, err
	}

	return fileCount, nil
}

func (l *LookThrough) LookForFiles(folderInspect bool) error {
	l.resetSkippedDirectories()
	defer l.printSkippedDirectories()

	if folderInspect {
		if err := l.loadExistingFiles(); err != nil {
			return fmt.Errorf("failed to load existing destination files: %w", err)
		}
	}

	absoluteNewPath, err := filepath.Abs(l.Vars.NewPath)
	if err != nil {
		return fmt.Errorf("failed to determine absolute destination path: %w", err)
	}
	absoluteNewPath = filepath.Clean(absoluteNewPath)

	if l.Vars.ShowProgressBar {
		fmt.Println("\n[+] COUNTING FILES...")
	}

	fileCount, err := l.countSourceFiles(absoluteNewPath)
	if err != nil {
		return fmt.Errorf("failed to count source files: %w", err)
	}

	l.Vars.FileCount = fileCount

	if fileCount == 0 {
		fmt.Print("\n[-] FAIL\n[-] NO FILES FOUND\n\n")
		return fmt.Errorf("no files found")
	}

	jobs := make(chan fileJob, l.workers*2)

	var workers sync.WaitGroup
	var workerError firstError

	progress, stopProgress, progressStopped := l.startProgressPrinter(fileCount)

	for worker := 0; worker < l.workers; worker++ {
		workers.Add(1)

		go func() {
			defer workers.Done()

			for job := range jobs {
				info, err := os.Stat(job.path)
				if err != nil {
					workerError.set(fmt.Errorf(
						"failed to fetch metadata for file '%s': %w",
						job.path,
						err,
					))
					progress.addProcessed()
					continue
				}

				if err := l.saveHash(job.name, job.path, info.Size()); err != nil {
					workerError.set(err)
				}

				progress.addProcessed()
			}
		}()
	}

	walkError := l.walkSourceFiles(absoluteNewPath, func(path string, entry fs.DirEntry) error {
		jobs <- fileJob{path: path, name: entry.Name()}
		return nil
	})

	close(jobs)
	workers.Wait()

	if stopProgress != nil {
		close(stopProgress)
		<-progressStopped
		fmt.Println()
	}

	if walkError != nil {
		return fmt.Errorf("failed to read source directory: %w", walkError)
	}
	if err := workerError.get(); err != nil {
		return err
	}

	if l.Vars.VerifyFiles {
		fmt.Println("\n[+] VERIFYING FILE HASHES...")

		if !l.verifyFiles() {
			fmt.Print("\n[-] FAIL\n[-] IN-MEMORY HASH CONSISTENCY CHECK FAILED\n\n")

			return fmt.Errorf("in-memory hash consistency check failed")
		}

		fmt.Println("[+] VERIFICATION PASSED")
	}

	totalSize := l.Vars.TotalSizeCount
	uniqueSize := l.Vars.SizeCount
	skippedSize, unit := getsize.GetSize(totalSize - uniqueSize)

	fmt.Print("\n[+] SUCCESS\n[+] UNIQUE FILES HAVE BEEN COPIED\n\n")
	fmt.Printf(">>> Source Files Scanned: %d\n", fileCount)
	fmt.Printf(">>> Unique Files Copied: %d\n", l.Vars.HashCount)
	fmt.Printf(">>> Duplicate Data Skipped: %.1f%s\n\n", skippedSize, unit)

	return nil
}

func (l *LookThrough) startProgressPrinter(total int64) (*progressState, chan struct{}, chan struct{}) {
	if !l.Vars.ShowProgressBar {
		return nil, nil, nil
	}

	progress := &progressState{total: total}
	stop := make(chan struct{})
	stopped := make(chan struct{})

	go func() {
		defer close(stopped)

		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()

		var lastProcessed int64 = -1

		printProgress := func() {
			processed, total := progress.snapshot()

			if processed != lastProcessed {
				progressbar.PrintProgressBar(processed, total)
				lastProcessed = processed
			}
		}

		printProgress()

		for {
			select {
			case <-ticker.C:
				printProgress()
			case <-stop:
				printProgress()
				return
			}
		}
	}()

	return progress, stop, stopped
}

func shouldIgnoreFile(fileName string) bool {
	switch strings.ToLower(fileName) {
	case "desktop.ini", "thumbs.db", ".ds_store":
		return true
	default:
		return false
	}
}
