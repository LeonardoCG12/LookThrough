package getpath

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func GetPath(pathArg string, cliArgs []string) (string, error) {
	var targetPath string

	if pathArg != "" {
		targetPath = pathArg
	} else if len(cliArgs) > 0 {
		targetPath = cliArgs[0]
	} else {
		fmt.Print("Choose a directory to look through: ")

		reader := bufio.NewReader(os.Stdin)

		for {
			input, _ := reader.ReadString('\n')
			targetPath = strings.TrimSpace(input)

			if targetPath != "" {
				break
			}

			fmt.Print("[-] path cannot be empty. Choose a directory: ")
		}
	}

	absPath, err := filepath.Abs(targetPath)
	if err != nil {
		return "", fmt.Errorf("failed to determine absolute path for '%s': %w", targetPath, err)
	}

	targetPath = absPath

	return targetPath, nil
}

func GetNewPath(targetPath string) string {
	baseName := filepath.Base(targetPath)

	return filepath.Join(targetPath, "new-"+baseName)
}
