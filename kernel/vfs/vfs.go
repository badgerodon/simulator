package vfs

import (
	"errors"
	"io"
	"os"
	"syscall"

	"github.com/gopherjs/gopherjs/js"
)

type (
	// A VFS is a virtual file system
	VFS struct {
		db      *js.Object
		files   map[int]vfsFile
		counter int
	}
	vfsFile interface {
		Close() error
		FCNTL(cmd int, arg int) (val int, err error)
		Read(p []byte) (n int, err error)
		Write(p []byte) (n int, err error)
	}
)

// New creates a new Virtual File System
func New() (*VFS, error) {
	type Result struct {
		vfs *VFS
		err error
	}

	c := make(chan Result, 1)
	req := js.Global.Get("indexedDB").Call("open", "vfs")
	req.Set("onerror", func(evt *js.Object) {
		c <- Result{
			err: errors.New(evt.String()),
		}
	})
	req.Set("onupgradeneeded", func(evt *js.Object) {
		db := evt.Get("target").Get("result")
		db.Call("createObjectStore", "files")
	})
	req.Set("onsuccess", func(evt *js.Object) {
		c <- Result{
			vfs: &VFS{
				db:      req.Get("result"),
				files:   make(map[int]vfsFile),
				counter: 10000,
			},
		}
	})

	result := <-c
	return result.vfs, result.err
}

// Close closes a file
func (vfs *VFS) Close(fd int) (err error) {
	f, ok := vfs.files[fd]
	if !ok {
		return nil
	}
	return f.Close()
}

// Open opens a file
func (vfs *VFS) Open(path string, mode int, perm uint32) (fd int, err error) {
	js.Global.Get("console").Call("log", "VFS", "Open", path, mode, perm)
	var f vfsFile
	switch path {
	case "/dev/null":
		f = &vfsInMemoryFile{
			vfs: vfs,

			path: path,
			mode: mode,
			perm: perm,
		}
	case "/dev/tty":
		f = new(ttyFile)
	default:
		imf := &vfsInMemoryFile{
			vfs: vfs,

			path: path,
			mode: mode,
			perm: perm,
		}

		type Result struct {
			res *js.Object
			err error
		}

		c := make(chan Result, 1)
		tx := vfs.db.Call("transaction", js.S{"files"}, "readonly")
		req := tx.Call("objectStore", "files").Call("get", path)
		req.Set("onsuccess", func(evt *js.Object) {
			c <- Result{
				res: evt.Get("target").Get("result"),
			}
		})
		req.Set("onerror", func(evt *js.Object) {
			c <- Result{
				err: errors.New(evt.Get("target").Get("error").String()),
			}
		})

		res := <-c
		if res.err != nil {
			return 0, err
		}

		if bs, ok := res.res.Interface().([]byte); ok && bs != nil {
			imf.data = bs
		} else {
			if mode&os.O_CREATE == 0 {
				return 0, os.ErrNotExist
			}
		}

		if mode&os.O_TRUNC > 0 {
			imf.position = 0
			imf.data = nil
		} else {
			imf.position = len(imf.data)
		}
		f = imf
	}

	fd = vfs.counter
	vfs.counter++

	vfs.files[fd] = f

	//js.Global.Get("console").Call("log", path, f.data)

	return fd, nil
}

// Read reads a file
func (vfs *VFS) Read(fd int, p []byte) (n int, err error) {
	f, ok := vfs.files[fd]
	if !ok {
		return 0, os.ErrInvalid
	}
	return f.Read(p)
}

// Write writes a file
func (vfs *VFS) Write(fd int, p []byte) (n int, err error) {
	// stderr/stdout
	if fd == 1 || fd == 2 {
		js.Global.Get("console").Call("log", string(p))
		return len(p), nil
	}

	f, ok := vfs.files[fd]
	if !ok {
		return 0, os.ErrInvalid
	}
	return f.Write(p)
}

// FCNTL calls FCNTL on a file
func (vfs *VFS) FCNTL(fd int, cmd int, arg int) (val int, err error) {
	f, ok := vfs.files[fd]
	if !ok {
		return -1, os.ErrInvalid
	}
	return f.FCNTL(cmd, arg)
}

type vfsInMemoryFile struct {
	vfs *VFS

	path string
	mode int
	perm uint32

	position int
	data     []byte //TODO: use blocks instead of a single slice of bytes
}

func (f *vfsInMemoryFile) Close() error {
	// nothing to do when we weren't open for writing
	if !canWrite(f.mode) {
		return nil
	}

	// flush the data to the db
	type Result struct {
		err error
	}
	c := make(chan Result, 1)
	tx := f.vfs.db.Call("transaction", js.S{"files"}, "readwrite")
	req := tx.Call("objectStore", "files").Call("put", f.data, f.path)
	req.Set("onsuccess", func(evt *js.Object) {
		c <- Result{}
	})
	req.Set("onerror", func(evt *js.Object) {
		c <- Result{
			err: errors.New(evt.Get("target").Get("error").String()),
		}
	})
	return (<-c).err
}

func (f *vfsInMemoryFile) FCNTL(cmd int, arg int) (val int, err error) {
	switch cmd {
	case syscall.F_GETFL:
		return f.mode, nil
	}

	return -1, syscall.EACCES
}

func (f *vfsInMemoryFile) Read(p []byte) (n int, err error) {
	if !canRead(f.mode) {
		return 0, os.ErrPermission
	}

	if f.position >= len(f.data) {
		return 0, io.EOF
	}

	if f.position+len(p) >= len(f.data) {
		copy(p, f.data[f.position:])
		n := len(f.data) - f.position
		f.position = len(f.data)
		return n, nil
	}

	copy(p, f.data[f.position:])
	f.position += len(p)
	return len(p), nil
}

func (f *vfsInMemoryFile) Write(p []byte) (n int, err error) {
	if !canWrite(f.mode) {
		return 0, os.ErrPermission
	}

	// make room
	diff := len(f.data[f.position:]) - len(p)
	if diff < 0 {
		f.data = append(f.data, make([]byte, -diff)...)
	}

	copy(f.data[f.position:], p)
	f.position += len(p)

	return len(p), nil
}

func canRead(mode int) bool {
	return mode&os.O_RDWR > 0 || mode&os.O_RDONLY > 0
}

func canWrite(mode int) bool {
	return mode&os.O_RDWR > 0 || mode&os.O_WRONLY > 0
}
