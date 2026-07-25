package main

import (
	"flag"
	"log"

	"github.com/LeonardoCG12/LookThrough/src/lookthrough"
	"github.com/LeonardoCG12/LookThrough/src/utils/gethardware"
	"github.com/LeonardoCG12/LookThrough/src/utils/getpath"
	"github.com/LeonardoCG12/LookThrough/src/utils/handlefolder"
	"github.com/LeonardoCG12/LookThrough/src/utils/variables"
)

func main() {
	barFlag := flag.Bool("b", false, "Enable visual progress bar")
	pathFlag := flag.String("p", "", "Target directory path to scan")
	sortFlag := flag.Bool("s", false, "Enable file sorting by extension")
	workersFlag := flag.Int("j", 0, "Number of concurrent workers (0 = automatic; manual range: 1 to 128)")
	verifyFlag := flag.Bool("verify", false, "Enable in-memory hash consistency check")
	safeCopyFlag := flag.Bool("safe-copy", false, "Copy through temporary files before finalizing destinations")
	ignoreDirectoryErrorsFlag := flag.Bool("ignore-directory-errors", false, "Skip unreadable source directories and report them instead of failing")

	flag.Parse()
	log.SetFlags(0)

	workers, err := gethardware.ResolveWorkerLimit(*workersFlag)
	if err != nil {
		log.Fatalf("[!] CRITICAL ERROR: %v\n", err)
	}

	targetPath, err := getpath.GetPath(*pathFlag, flag.Args())
	if err != nil {
		log.Fatalf("[!] CRITICAL ERROR: %v\n", err)
	}

	newPath := getpath.GetNewPath(targetPath)

	folderInspect, err := handlefolder.MakeNewDir(newPath)
	if err != nil {
		log.Fatalf("[!] CRITICAL ERROR: %v\n", err)
	}

	vars := variables.LookThroughVars{
		MyPath:                targetPath,
		NewPath:               newPath,
		IgnoreDirectoryErrors: *ignoreDirectoryErrorsFlag,
		SafeCopy:              *safeCopyFlag,
		ShowProgressBar:       *barFlag,
		SortFile:              *sortFlag,
		VerifyFiles:           *verifyFlag,
		Workers:               workers,
	}

	if err := lookthrough.NewLookThrough(vars).LookForFiles(folderInspect); err != nil {
		log.Fatalf("[!] CRITICAL ERROR: %v\n", err)
	}
}
