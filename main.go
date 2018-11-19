package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"
)

const gitExecutable = "git"

var count int

func fileExists(path string) bool {
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

func chdir(path string) {
	if err := os.Chdir(path); err != nil {
		panic(err)
	}
}

func gitDirty(path string) {
	wd, _ := os.Getwd()
	defer chdir(wd)
	chdir(path)
	output, err := exec.Command(gitExecutable, "status", "--short").CombinedOutput()
	if err != nil {
		fmt.Printf("%s\n", output)
		logrus.Fatal(err)
	}
	if len(output) > 0 {
		fmt.Println(path)
		for _, line := range strings.Split(strings.TrimSpace(string(output)), "\n") {
			fmt.Printf("  %s\n", line)
		}
		count++
	}
}

func walk(path string, info os.FileInfo, err error) error {
	if err != nil {
		logrus.Warnf("error walking %s: %v", path, err)
		return filepath.SkipDir
	}
	if info.IsDir() && info.Name() == ".git" {
		basepath, err := filepath.Abs(filepath.Dir(path))
		if err != nil {
			logrus.Fatalf("error getting absolute directory for %s: %v", path, err)
		}
		if fileExists(filepath.Join(path, "HEAD")) && fileExists(filepath.Join(path, "refs")) {
			gitDirty(basepath)
		}
		return filepath.SkipDir
	}
	return nil
}

func init() {
	_, err := exec.LookPath(gitExecutable)
	if err != nil {
		logrus.Fatalf("%s not present in $PATH", gitExecutable)
	}
	flag.Parse()
}

func main() {
	args := flag.Args()
	if len(flag.Args()) == 0 {
		args = append(args, ".")
	}
	for _, arg := range args {
		filepath.Walk(arg, walk)
	}
	if count > 0 {
		os.Exit(1)
	}
}
