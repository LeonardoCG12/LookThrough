package create

import (
	"fmt"
	"os"
)

func MakeNewDir(newPath string) error {
	if err := os.Mkdir(newPath, 0755); err != nil {
		return fmt.Errorf("erro ao criar diretório %s: %w", newPath, err)
	}
	return nil
}
