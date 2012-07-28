package archive

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type Device interface {
	Open(path string) (rc io.ReadCloser, err error)
}

type fsDevice string

// FsDevice returns an archive device that represents a native filesystem
// path.
func FsDevice(rootPath string) (fd Device, err error) {
	path, err := filepath.Abs(rootPath)
	return fsDevice(path), err
}

func (fd fsDevice) Open(path string) (rc io.ReadCloser, err error) {
	return os.Open(filepath.Join(string(fd), path))
}

type zipDevice struct{ reader *zip.Reader }

// FileZipDevice returns an archive device that represents the contents of a
// zip file.
func FileZipDevice(path string) (zd Device, err error) {
	r, err := zip.OpenReader(path)
	if err != nil {
		return
	}
	zd = zipDevice{&r.Reader}
	return
}

func (zd zipDevice) Open(path string) (rc io.ReadCloser, err error) {
	for _, f := range zd.reader.File {
		if f.Name == path {
			return f.Open()
		}
	}
	err = errors.New(fmt.Sprintf("File '%s' not found", path))
	return
}

type multiDevice []Device

func (md multiDevice) Open(path string) (rc io.ReadCloser, err error) {
	for _, d := range ([]Device)(md) {
		rc, err = d.Open(path)
		if err == nil {
			return
		}
	}
	return
}

func New(devs ...Device) Device {
	return multiDevice(devs)
}
