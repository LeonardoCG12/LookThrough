package main

import (
	"crypto/md5"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

var (
	fileCount   int
	hashCount   int
	hashList    []string
	hashListAll []string
	home, _     = os.UserHomeDir()
	num         string
	mem         = map[string]int{}
	newPath     string
	myPath      string
	sep         = fmt.Sprintf("%c", os.PathSeparator)
)

func makeNewDir() {

	if myPath != "" {
		return
	}

	refactorPath()

	myPath = strings.TrimSuffix(myPath, sep)
	arr := strings.Split(myPath, sep)
	newPath = fmt.Sprintf("%s/new-%s", myPath, arr[len(arr)-1])

	os.Mkdir(newPath, 0755)
}

func refactorPath() string {

	if len(os.Args) > 1 {
		myPath = os.Args[1]
	} else {
		fmt.Print("Choose a directory to look through: ")
		fmt.Scanf("%s", &myPath)
	}

	myPath = strings.Replace(myPath, "~", home, -1)

	return myPath
}

func getMD5Checksum(filePath string, fileName string) {
	file, err := ioutil.ReadFile(filePath)

	if err != nil {
		log.Fatal(err)
	}

	hasher := md5.New()
	hasher.Write(file)

	if err != nil {
		log.Fatal(err)
	}

	checksum := fmt.Sprintf("%x", hasher.Sum(nil))

	saveHash(fileName, file, checksum)
}

func saveHash(fileName string, file []byte, md5Sum string) {
	value, isThere := mem[fileName]
	look4Hash := look4Hashes(fileName, md5Sum)

	if look4Hash == 1 {
		hashCount += 1

		if isThere {
			mem[fileName] += 1
			num = fmt.Sprintf("%d", value+1)
		} else {
			num = ""
		}

		hashList = append(hashList, fileName, md5Sum)
		arr := strings.Split(fileName, ".")
		filePath := fmt.Sprintf("%s%s%s(%s).%s", newPath, sep, arr[0], num, arr[len(arr)-1])
		err := ioutil.WriteFile(filePath, file, 0644)

		if err != nil {
			log.Fatal(err)
		}

	} else if look4Hash == 2 {
		hashCount += 1
		mem[fileName] = 0

		hashList = append(hashList, fileName, md5Sum)
		filePath := fmt.Sprintf("%s%s%s", newPath, sep, fileName)
		err := ioutil.WriteFile(filePath, file, 0644)

		if err != nil {
			log.Fatal(err)
		}

	}

	hashListAll = append(hashListAll, fileName, md5Sum)

}

func look4Hashes(fileName string, md5Sum string) int {

	for _, element := range hashList {

		if md5Sum == element {
			return 0
		}

	}

	for _, element := range hashList {

		if fileName == element {
			return 1
		}

	}

	return 2
}

func look4Files() {
	arr := strings.Split(newPath, sep)
	newPathDir := arr[len(arr)-1]

	err := filepath.Walk(myPath, func(path string, info os.FileInfo, err error) error {
		arr = strings.Split(path, sep)
		checkDir := arr[len(arr)-2]

		if err != nil {
			log.Fatal(err)
		}

		if !info.IsDir() && checkDir != newPathDir {
			fileCount += 1
			getMD5Checksum(path, info.Name())
		}

		return nil
	})

	if err != nil {
		log.Fatal(err)
	}

	if verifyFiles() {
		fmt.Print("\n[+] SUCCESS\n")
		fmt.Print("[+] ALL FILES HAVE BEEN COPIED\n")
		fmt.Printf("\n>>> Old Files: %d\n", fileCount)
		fmt.Printf(">>> New Files: %d\n\n", hashCount)
	} else {
		fmt.Print("\n[-] FAIL\n")
		fmt.Print("[-] SOMETHING WENT WRONG\n\n")
	}

}

func verifyFiles() bool {
	var integrity bool

start:

	for i := 1; i < len(hashListAll); i += 2 {

		for j := 1; j < len(hashList); j += 2 {

			if hashListAll[i] == hashList[j] {
				integrity = true
				continue start
			} else {
				integrity = false
			}

		}

	}

	return integrity
}

func main() {
	makeNewDir()
	look4Files()
}
