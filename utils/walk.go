package utils

import (
	"os"
	"path"
)

func Walk(dir string, callback func(root string, dirs []string, files []string) error) error {
	if !path.IsAbs(dir) {
		wd, err := os.Getwd()
		if err != nil {
			return err
		}
		dir = path.Join(wd, dir)
	}

	dirFiles, err := os.ReadDir(dir)

	if err != nil {
		return err
	}

	var dirs []string
	var files []string

	for _, file := range dirFiles {
		if file.IsDir() {
			dirs = append(dirs, file.Name())
		} else {
			files = append(files, file.Name())
		}
	}

	err = callback(dir, dirs, files)

	if err != nil {
		return err
	}

	for _, subdir := range dirs {
		nextDir := path.Join(dir, subdir)
		err := Walk(nextDir, callback)

		if err != nil {
			return err
		}
	}

	return nil
}
