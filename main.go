package main

import (
	"github.com/LeonardoCG12/LookThrough/create"
	"github.com/LeonardoCG12/LookThrough/lookthrough"
	"github.com/LeonardoCG12/LookThrough/utils/gethome"
	"github.com/LeonardoCG12/LookThrough/utils/getpath"
	"github.com/LeonardoCG12/LookThrough/utils/getseparator"
	"github.com/LeonardoCG12/LookThrough/variables"
)

func main() {
	home, _ := gethome.GetHome()
	getSeparator := getseparator.GetSeparator()
	getPath := getpath.GetPath(home, getSeparator)
	getNewPath := getpath.GetNewPath(getPath, getSeparator)

	vars := variables.LookThroughVars{
		FileCount:      0,
		HashCount:      0,
		HashList:       []variables.FileHash{},
		HashListAll:    []variables.FileHash{},
		Mem:            map[string]int{},
		MyPath:         getPath,
		NewPath:        getNewPath,
		Num:            "",
		Separator:      getSeparator,
		SizeCount:      0,
		TotalSizeCount: 0,
	}

	create.MakeNewDir(getNewPath)
	lookthrough.NewLookThrough(vars).LookForFiles()
}
