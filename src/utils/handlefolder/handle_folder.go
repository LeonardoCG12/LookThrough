package handlefolder

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func MakeNewDir(newPath string) error {
	info, err := os.Stat(newPath)

	if err == nil && info.IsDir() {
		reader := bufio.NewReader(os.Stdin)

		for {
			fmt.Printf("\n[!] The folder '%s' already exists.\n", newPath)
			fmt.Print("[?] Do you want to delete its content or amend it? (d/a): ")

			input, _ := reader.ReadString('\n')
			input = strings.TrimSpace(strings.ToLower(input))

			if input == "d" || input == "delete" {
				fmt.Printf("[+] Deleting existing contents...\n\n")

				if err := os.RemoveAll(newPath); err != nil {
					return fmt.Errorf("error deleting directory: %w", err)
				}

				break
			} else if input == "a" || input == "amend" {
				fmt.Printf("[+] Amending existing directory...\n\n")

				return nil
			} else {
				fmt.Println("[-] Invalid option. Please use 'd' (delete) or 'a' (amend).")
			}
		}
	}

	if err := os.MkdirAll(newPath, 0755); err != nil {
		return fmt.Errorf("error creating directory: %w", err)
	}

	return nil
}
