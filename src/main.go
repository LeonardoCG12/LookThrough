package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/LeonardoCG12/LookThrough/src/lookthrough"
	"github.com/LeonardoCG12/LookThrough/src/utils/getpath"
	"github.com/LeonardoCG12/LookThrough/src/utils/handlefolder"
	"github.com/LeonardoCG12/LookThrough/src/utils/variables"
)

func main() {
	barFlag := flag.Bool("b", false, "Enable visual progress bar")
	pathFlag := flag.String("p", "", "Target directory path to scan")
	sortFLag := flag.Bool("s", false, "Enable file sorting by extension")

	flag.Parse()

	getPath, err := getpath.GetPath(*pathFlag, flag.Args())

	if err != nil {
		fmt.Printf("[!] CRITICAL ERROR: %v\n", err)
		os.Exit(1)
	}

	getNewPath := getpath.GetNewPath(getPath)

	folderInspect, err := handlefolder.MakeNewDir(getNewPath)

	if err != nil {
		fmt.Printf("[!] CRITICAL ERROR: %v\n", err)
		os.Exit(1)
	}

	vars := variables.LookThroughVars{
		FileCount:       0,
		HashCount:       0,
		HashList:        []variables.FileHash{},
		HashListAll:     []variables.FileHash{},
		HashMap:         make(map[string]bool),
		NameMap:         make(map[string]bool),
		Mem:             map[string]int{},
		MyPath:          getPath,
		NewPath:         getNewPath,
		Num:             "",
		ShowProgressBar: *barFlag,
		SizeCount:       0,
		SortFile:        *sortFLag,
		TotalSizeCount:  0,
	}

	if err := lookthrough.NewLookThrough(vars).LookForFiles(folderInspect); err != nil {
		fmt.Printf("[!] CRITICAL ERROR: %v\n", err)
		os.Exit(1)
	}
}
