package create

import (
	"os"
)

func MakeNewDir(newPath string) {

	os.Mkdir(newPath, 0755)
}
