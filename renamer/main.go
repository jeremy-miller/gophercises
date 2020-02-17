package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type file struct {
	name string
	path string
}

func main() {
	dirName := "sample"
	var toRename []file
	filepath.Walk(dirName, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() { // we're not renaming directories, but don't return error since then it won't recurse into the directory
			return nil
		}
		if _, err := match(info.Name()); err == nil {
			toRename = append(toRename, file{
				name: info.Name(),
				path: path,
			})
		}
		return nil
	})
	for _, f := range toRename {
		fmt.Printf("%q\n", f)
	}

	for _, orig := range toRename {
		var n file
		var err error
		n.name, err = match(orig.name)
		if err != nil {
			fmt.Println("Error matching:", orig.path, err.Error())
		}
		n.path = filepath.Join(dirName, n.name)
		fmt.Printf("mv %s => %s\n", orig.path, n.path)
		err = os.Rename(orig.path, n.path)
		if err != nil {
			fmt.Println("Error renaming:", orig.path, err.Error())
		}
	}
}

// match returns the new file name, or an error if the file name didn't match the pattern.
func match(filename string) (string, error) {
	pieces := strings.Split(filename, ".")
	ext := pieces[len(pieces)-1]                              // make sure we get the extension (in case filename has a period in them)
	tmpFilename := strings.Join(pieces[0:len(pieces)-1], ".") // add back periods to filename
	pieces = strings.Split(tmpFilename, "_")
	number, err := strconv.Atoi(pieces[len(pieces)-1])
	if err != nil {
		return "", fmt.Errorf("%s did not match the pattern", filename)
	}
	name := strings.Join(pieces[0:len(pieces)-1], "_")
	return fmt.Sprintf("%s - %d.%s", strings.Title(name), number, ext), nil
}
