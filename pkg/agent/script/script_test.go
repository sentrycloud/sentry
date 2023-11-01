package script

import (
	"fmt"
	"testing"
)

func TestFileScanner(t *testing.T) {
	var fileScanner = FileScanner{}
	scripts := fileScanner.GetAllFiles("../script")
	for _, script := range scripts {
		fmt.Printf("%s\n", script)
	}
}
