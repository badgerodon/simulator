package main

import (
	"io/ioutil"
	"log"

	"syscall"
)

func init() {
	log.SetFlags(0)

	vfs, err := NewVFS()
	if err != nil {
		log.Fatal(err)
	}

	syscall.DefaultCloseFunction = func(fd int) error {
		return vfs.Close(fd)
	}

	syscall.DefaultOpenFunction = func(path string, mode int, perm uint32) (fd int, err error) {
		return vfs.Open(path, mode, perm)
	}

	syscall.DefaultWriteFunction = func(fd int, p []byte) (int, error) {
		return vfs.Write(fd, p)
	}

	syscall.DefaultFCNTLFunction = func(fd int, cmd int, arg int) (val int, err error) {
		return vfs.FCNTL(fd, cmd, arg)
	}
}

func main() {
	log.Println("(1)", readFile())
	log.Println("(2)", writeFile())
	log.Println("(3)", readFile())
}

func writeFile() error {
	bs := []byte("Hello World")
	err := ioutil.WriteFile("/tmp/example.txt", bs, 0666)
	if err != nil {
		return err
	}

	log.Println("WROTE", string(bs))

	return nil
}

func readFile() error {
	bs, err := ioutil.ReadFile("/tmp/example.txt")
	if err != nil {
		return err
	}

	log.Println("READ", string(bs))

	return nil
}
