package getseparator

import (
	"os"
)

func GetSeparator() string {
	return string(os.PathSeparator)
}
