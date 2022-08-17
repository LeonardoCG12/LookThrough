package lookthrough

import (
	"crypto/md5"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/LeonardoCG12/Look-Through-Go/variables"
)

const (
	_  = iota
	KB = 1 << (iota * 10)
	MB
	GB
	TB
)

type LookThrough struct {
	LookThrough variables.LookThroughVars
}

func NewLookThrough(lookThrough variables.LookThroughVars) *LookThrough {
	return &LookThrough{LookThrough: lookThrough}
}

func (l *LookThrough) getMD5Checksum(filePath string, fileName string, fileSize int64) {
	fin, err := os.Open(filePath)

	if err != nil {
		log.Fatal(err)
	}

	defer fin.Close()

	hasher := md5.New()

	_, err = io.Copy(hasher, fin)

	if err != nil {
		log.Fatal(err)
	}

	checksum := fmt.Sprintf("%x", hasher.Sum(nil))

	l.saveHash(fileName, fileSize, filePath, checksum)
}

func (l *LookThrough) saveHash(fileName string, fileSize int64, filePath string, md5Sum string) {
	value, isThere := l.LookThrough.Mem[fileName]
	lookForHash := l.lookForHashes(fileName, md5Sum)

	if lookForHash == 1 {
		l.LookThrough.HashCount += 1

		l.saveSize(true, fileSize)

		if isThere {
			l.LookThrough.Mem[fileName] += 1
			l.LookThrough.Num = fmt.Sprintf("%d", value+1)
		} else {
			l.LookThrough.Num = ""
		}

		l.LookThrough.HashList = append(l.LookThrough.HashList, fileName, md5Sum)
		arr := strings.Split(fileName, ".")
		newFilePath := fmt.Sprintf("%s%s%s(%s).%s", l.LookThrough.NewPath, l.LookThrough.Separator, arr[0], l.LookThrough.Num, arr[len(arr)-1])
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

	} else if lookForHash == 2 {
		l.LookThrough.HashCount += 1
		l.LookThrough.Mem[fileName] = 0

		l.saveSize(true, fileSize)

		l.LookThrough.HashList = append(l.LookThrough.HashList, fileName, md5Sum)
		newFilePath := fmt.Sprintf("%s%s%s", l.LookThrough.NewPath, l.LookThrough.Separator, fileName)
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

	l.LookThrough.HashListAll = append(l.LookThrough.HashListAll, fileName, md5Sum)

	l.saveSize(false, fileSize)
}

func (l *LookThrough) lookForHashes(fileName string, md5Sum string) int {

	for _, element := range l.LookThrough.HashList {

		if md5Sum == element {
			return 0
		}

	}

	for _, element := range l.LookThrough.HashList {

		if fileName == element {
			return 1
		}

	}

	return 2
}

func (l *LookThrough) LookForFiles() {
	arr := strings.Split(l.LookThrough.NewPath, l.LookThrough.Separator)
	newPathDir := arr[len(arr)-1]

	err := filepath.Walk(l.LookThrough.MyPath, func(path string, info os.FileInfo, err error) error {
		arr = strings.Split(path, l.LookThrough.Separator)
		checkDir := arr[len(arr)-2]

		if err != nil {
			log.Fatal(err)
		}

		if !info.IsDir() && checkDir != newPathDir {
			l.LookThrough.FileCount += 1
			l.getMD5Checksum(path, info.Name(), info.Size())
		}

		return nil
	})

	if err != nil {
		log.Fatal(err)
	}

	if l.verifyFiles() {
		size, unit := getSize(l.LookThrough.TotalSizeCount - l.LookThrough.SizeCount)

		fmt.Print("\n[+] SUCCESS\n")
		fmt.Print("[+] ALL FILES HAVE BEEN COPIED\n\n")
		fmt.Printf(">>> Old Files: %d\n", l.LookThrough.FileCount)
		fmt.Printf(">>> New Files: %d\n", l.LookThrough.HashCount)
		fmt.Printf(">>> Freed Storage: %.1f%s\n\n", size, unit)
	} else {
		fmt.Print("\n[-] FAIL\n")
		fmt.Print("[-] SOMETHING WENT WRONG\n\n")
	}

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

func (l *LookThrough) saveSize(isNew bool, fileSize int64) {

	if isNew {
		l.LookThrough.SizeCount += fileSize
	} else {
		l.LookThrough.TotalSizeCount += fileSize
	}

}

func (l *LookThrough) verifyFiles() bool {
	var integrity bool

start:

	for i := 1; i < len(l.LookThrough.HashListAll); i += 2 {

		for j := 1; j < len(l.LookThrough.HashList); j += 2 {

			if l.LookThrough.HashListAll[i] == l.LookThrough.HashList[j] {
				integrity = true
				continue start
			} else {
				integrity = false
			}

		}

	}

	return integrity
}
