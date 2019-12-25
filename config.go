package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

const (
	dotDirName = ".runstatic"
	envDotDir  = "RUNSTATIC_DIR"
)

type dotDir string
type dotFile string

func (d dotDir) File(path ...string) dotFile {
	return dotFile(filepath.Join(append([]string{string(d)}, path...)...))
}

func (d dotFile) Path() string { return string(d) }
func (d dotFile) Read() (string, error) {
	b, err := ioutil.ReadFile(string(d))
	return string(b), err
}
func (d dotFile) Write(s string) error {
	if err := os.MkdirAll(filepath.Dir(string(d)), 0700); err != nil {
		return err
	}
	return ioutil.WriteFile(string(d), []byte(s), 0644)
}

func (d dotFile) Exists() (bool, error) {
	fi, err := os.Stat(string(d))
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	if !fi.Mode().IsRegular() {
		return true, fmt.Errorf("file exists, but not regular: %s", fi.Mode())
	}
	return true, nil
}

func mustDotDir() dotDir {
	if v := os.Getenv(envDotDir); v != "" {
		return dotDir(v)
	}
	home := os.Getenv("HOME")
	if home == "" {
		home = os.Getenv("USERPROFILE")
	}
	if home == "" {
		panic("home directory cannot be inferred")
	}
	return dotDir(filepath.Join(home, dotDirName))
}
