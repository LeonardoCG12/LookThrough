package getpath

import (
	"fmt"
	"path/filepath"
	"strings"
)

func GetPath(homeDir, separator, pathArg string, cliArgs []string) string {
	var targetPath string

	if pathArg != "" {
		targetPath = pathArg
	} else if len(cliArgs) > 0 {
		targetPath = cliArgs[0]
	} else {
		fmt.Print("Choose a directory to look through: ")
		fmt.Scanf("%s", &targetPath)
	}

	absPath, err := filepath.Abs(targetPath)
	if err == nil {
		targetPath = absPath
	}

	return targetPath
}

func GetNewPath(targetPath, separator string) string {
	baseName := filepath.Base(targetPath)
	return filepath.Join(targetPath, "new-"+baseName)
}

func GetNewFilePath(newPath, separator, fileName, fileNumber string, status int) string {
	if status == 1 {
		fileExtension := filepath.Ext(fileName)
		baseName := strings.TrimSuffix(fileName, fileExtension)
		newFileName := fmt.Sprintf("%s(%s)%s", baseName, fileNumber, fileExtension)

		return filepath.Join(newPath, newFileName)
	}

	return filepath.Join(newPath, fileName)
}
