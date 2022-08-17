package gethome

import (
	"log"
	"os"
)

func GetHome() string {
	home, err := os.UserHomeDir()

	if err != nil {
		log.Fatal(err)
	}

	return home
}
