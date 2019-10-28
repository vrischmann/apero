// +build !release

package ui

import "io/ioutil"

// GetFile returns the content of the file at path
func GetFile(path string) ([]byte, error) {
	if path[0] != '/' {
		panic("path should start with /")
	}

	path = "./ui" + path

	return ioutil.ReadFile(path)
}
