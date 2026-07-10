package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/LeonardoCG12/LookThrough/lookthrough"
	"github.com/LeonardoCG12/LookThrough/utils/create"
	"github.com/LeonardoCG12/LookThrough/utils/gethome"
	"github.com/LeonardoCG12/LookThrough/utils/getpath"
	"github.com/LeonardoCG12/LookThrough/utils/getseparator"
	"github.com/LeonardoCG12/LookThrough/variables"
)

func main() {
	barFlag := flag.Bool("b", false, "Enable visual progress bar")
	pathFlag := flag.String("p", "", "Target directory path to scan")

	flag.Parse()

	home, err := gethome.GetHome()
	if err != nil {
		fmt.Printf("Critical error getting Home directory: %v\n", err)
		os.Exit(1)
	}

	getSeparator := getseparator.GetSeparator()

	getPath := getpath.GetPath(home, getSeparator, *pathFlag, flag.Args())
	getNewPath := getpath.GetNewPath(getPath, getSeparator)

	vars := variables.LookThroughVars{
		FileCount:      0,
		HashCount:      0,
		HashList:       []variables.FileHash{},
		HashListAll:    []variables.FileHash{},
		HashMap:        make(map[string]bool),
		NameMap:        make(map[string]bool),
		Mem:            map[string]int{},
		MyPath:         getPath,
		NewPath:        getNewPath,
		Num:            "",
		Separator:      getSeparator,
		SizeCount:      0,
		TotalSizeCount: 0,
	}

	create.MakeNewDir(getNewPath)

	if err := lookthrough.NewLookThrough(vars).LookForFiles(*barFlag); err != nil {
		fmt.Printf("\n[-] CRITICAL ERROR: %v\n", err)
	}
}
