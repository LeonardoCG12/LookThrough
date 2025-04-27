package getpath

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func GetPath(home, separator string) string {
	var myPath string

	if len(os.Args) > 1 {
		myPath = os.Args[1]
	} else {
		fmt.Print("Choose a directory to look through: ")
		fmt.Scanf("%s", &myPath)
	}

	myPath = strings.Replace(myPath, "~", home, -1)
	myPath = strings.TrimSuffix(myPath, separator)
	return myPath
}

func GetNewPath(myPath, separator string) string {
	base := filepath.Base(myPath)
	return filepath.Join(myPath, "new-"+base)
}

func GetNewFilePath(newPath, separator, fileName, num string, status int) string {
	if status == 1 {
		ext := filepath.Ext(fileName)
		name := strings.TrimSuffix(fileName, ext)
		newName := fmt.Sprintf("%s(%s)%s", name, num, ext)
		return filepath.Join(newPath, newName)
	}
	return filepath.Join(newPath, fileName)
}
