package getpath

import (
	"fmt"
	"os"
	"strings"
)

func GetPath(home string, separator string) string {
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

func GetNewPath(myPath string, separator string) string {
	var newPath string

	myPath = strings.TrimSuffix(myPath, separator)
	arr := strings.Split(myPath, separator)
	newPath = fmt.Sprintf("%s/new-%s", myPath, arr[len(arr)-1])

	return newPath
}
