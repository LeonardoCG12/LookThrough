package getpath

import (
	"fmt"
	"os"
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
	var newPath string

	myPath = strings.TrimSuffix(myPath, separator)
	arr := strings.Split(myPath, separator)
	newPath = fmt.Sprintf("%s%snew-%s", myPath, separator, arr[len(arr)-1])

	return newPath
}

func GetNewFilePath(newPath, separator, fileName, num string, status int) string {

	if status == 1 {
		arr := strings.Split(fileName, ".")
		newFilePath := fmt.Sprintf("%s%s%s(%s).%s", newPath, separator, arr[0], num, arr[len(arr)-1])

		return newFilePath
	}

	newFilePath := fmt.Sprintf("%s%s%s", newPath, separator, fileName)

	return newFilePath
}
