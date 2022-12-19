package utils

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"strings"
)

func CheckPathExist(path string) bool {
	filename := strings.Split(path, "/")

	var outbuf1 bytes.Buffer
	var outbuf2 bytes.Buffer

	find := exec.Command("find", "-wholename", path)
	grep := exec.Command("grep", filename[len(filename)-2])

	find.Stdout = &outbuf1
	grep.Stdin = &outbuf1
	grep.Stdout = &outbuf2

	err := find.Run()
	if err != nil {
		fmt.Println("find " + err.Error())
		return false
	}
	// fmt.Println(outbuf1.String())
	err = grep.Run()
	if err != nil {
		// fmt.Println("grep " + err.Error())
		return false
	}
	fmt.Println(outbuf2.String())
	return true
}

func CreatePath(path string) error {
	if !CheckPathExist(path) {
		err := os.MkdirAll(path, fs.ModePerm)
		if err != nil {
			fmt.Println(err.Error())
			return err
		}
		return nil
	}
	return nil
}
