package vfs

import (
	"os"
)

type ttyFile struct {
}

func (f *ttyFile) Close() error {
	return nil
}

func (f *ttyFile) FCNTL(cmd int, arg int) (val int, err error) {
	return 0, nil
}

func (f *ttyFile) Read(p []byte) (n int, err error) {
	return os.Stdin.Read(p)
}

func (f *ttyFile) Write(p []byte) (n int, err error) {
	return os.Stdout.Write(p)
}
