package main

import (
	"fmt"
	"path/filepath"
)

// getTemplates variadic function that takes any number of single glob patterns
func getTemplates(paths ...string) (templates []string, err error) {
	for _, path := range paths {
		files, err := filepath.Glob(path)
		if err != nil {
			return nil, fmt.Errorf("error reading templates from this path: %v. Message: %v", path, err)
		}
		templates = append(templates, files...)
	}

	return templates, nil
}
