package main

import (
	"fmt"
	"io/fs"

	"github.com/aerostackdev/cli/internal/templates"
)

func main() {
	err := fs.WalkDir(templates.FS, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		fmt.Println(path)
		return nil
	})
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}
}
