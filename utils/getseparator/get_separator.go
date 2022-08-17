package getseparator

import (
	"fmt"
	"os"
)

func GetSeparator() string {
	separator := fmt.Sprintf("%c", os.PathSeparator)

	return separator
}
