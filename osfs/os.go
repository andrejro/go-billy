// Package osfs provides a billy filesystem for the OS.
package osfs // import "gopkg.in/src-d/go-billy.v4/osfs"

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"

	"gopkg.in/src-d/go-billy.v4"
	"gopkg.in/src-d/go-billy.v4/helper/chroot"
)

const (
	defaultDirectoryMode = 0755
	defaultCreateMode    = 0666
)

// OS is a filesystem based on the os filesystem.
type OS struct{}

// New returns a new OS filesystem.
func New(baseDir string) billy.Filesystem {
	return chroot.New(&OS{}, baseDir)
}

func (fs *OS) Create(filename string) (billy.File, error) {
	return fs.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_TRUNC, defaultCreateMode)
}

func (fs *OS) OpenFile(filename string, flag int, perm os.FileMode) (billy.File, error) {
	if flag&os.O_CREATE != 0 {
		if err := fs.createDir(filename); err != nil {
			return nil, err
		}
	}

	f, err := os.OpenFile(filename, flag, perm)
	if err != nil {
		return nil, err
	}
	return &file{File: f}, err
}

func (fs *OS) createDir(fullpath string) error {
	dir := filepath.Dir(fullpath)
	if dir != "." {
		if err := os.MkdirAll(dir, defaultDirectoryMode); err != nil {
			return err
		}
	}

	return nil
}

func (fs *OS) ReadDir(path string) ([]os.FileInfo, error) {
	l, err := ioutil.ReadDir(path)
	if err != nil {
		return nil, err
	}

	var s = make([]os.FileInfo, len(l))
	for i, f := range l {
		s[i] = f
	}

	return s, nil
}

func (fs *OS) Rename(from, to string) error {
	if err := fs.createDir(to); err != nil {
		return err
	}

	return os.Rename(from, to)
}

func (fs *OS) MkdirAll(path string, perm os.FileMode) error {
	return os.MkdirAll(path, defaultDirectoryMode)
}

func (fs *OS) Open(filename string) (billy.File, error) {
	return fs.OpenFile(filename, os.O_RDONLY, 0)
}

func (fs *OS) Remove(filename string) error {
	return os.Remove(filename)
}

func (fs *OS) TempFile(dir, prefix string) (billy.File, error) {
	if err := fs.createDir(dir + string(os.PathSeparator)); err != nil {
		return nil, err
	}

	f, err := ioutil.TempFile(dir, prefix)
	if err != nil {
		return nil, err
	}
	return &file{File: f}, nil
}

func (fs *OS) Join(elem ...string) string {
	return filepath.Join(elem...)
}

func (fs *OS) RemoveAll(path string) error {
	return os.RemoveAll(filepath.Clean(path))
}

func (fs *OS) Lstat(filename string) (os.FileInfo, error) {
	return os.Lstat(filepath.Clean(filename))
}

func (fs *OS) Symlink(target, link string) error {
	if err := fs.createDir(link); err != nil {
		return err
	}

	return os.Symlink(target, link)
}

func (fs *OS) Readlink(link string) (string, error) {
	return os.Readlink(link)
}

// file is a wrapper for an os.File which adds support for file locking.
type file struct {
	*os.File
	m sync.Mutex
}
