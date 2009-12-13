// Archive and filesystem utilities

package fs

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"container/vector"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"os"
)

type Archive struct {
	tarball []byte
}

func ArchiveFromTarGzFile(filename string) (result *Archive, err os.Error) {
	tarball, e := UnpackGzFile(filename)
	if e != nil {
		err = e
		return
	}
	result = NewTarArchive(tarball)
	return
}

func NewTarArchive(tarData []byte) (result *Archive) {
	result = new(Archive)
	// XXX: Stupid verbose array copy
	result.tarball = make([]byte, len(tarData))
	for i, v := range tarData {
		result.tarball[i] = v
	}

	return
}

func (self *Archive)ReadFile(name string) (data []byte, err os.Error) {
	// XXX: Should we use some kind of caching here?
	tr := tar.NewReader(bytes.NewBuffer(self.tarball))
	for {
		header, e := tr.Next()
		if e != nil {
			err = e
			return
		}
		if header == nil {
			err = os.NewError(fmt.Sprintf("File '%s' not found in archive.", name))
			return
		}
		if header.Name == name {
// ReadAll doesn't work with tar?
//			data, err = ioutil.ReadAll(tr)
			data = make([]byte, header.Size)
			n, err := io.ReadFull(tr, data)
			if err == nil && int64(n) != header.Size {
				err = os.NewError(fmt.Sprintf(
					"File was %v bytes, but read %v bytes",
					header.Size, n))
			}
			break
		}
	}
	return
}

func (self *Archive)ListFiles() (list []string, err os.Error) {
	names := new(vector.StringVector)
	tr := tar.NewReader(bytes.NewBuffer(self.tarball))
	for {
		header, e := tr.Next()
		if e != nil {
			err = e
			return
		}
		if header == nil {
			break
		}
		names.Push(header.Name)
	}

	list = names.Data()
	return
}

// Returns a file name that can be used to access the currently run binary.
func SelfExe() string {
	// XXX: Works only on Linux
	return "/proc/self/exe"
}

func UnpackGzFile(filename string) (data []byte, err os.Error) {
	fileData, e := ioutil.ReadFile("/proc/self/exe")
	if e != nil {
		err = e
		return
	}
	return UnpackGz(fileData)
}

func UnpackGz(fileData []byte) (data []byte, err os.Error) {
	// The third character of the magic string is not an official part of
	// the gz file magic sequence. It indicates compression method
	// deflate, which seems to be used in most gz files. This is great for
	// decreasing false positives, but it may fail with exotic gz files.
	gzMagic, _ := hex.DecodeString("1f8b08")

	sites := magicSites(fileData, gzMagic)

	for _, n := range sites {
		inf, e1 := gzip.NewInflater(bytes.NewBuffer(fileData[n:]))
		if e1 != nil { continue } // It wasn't really gzip data.
		unpacked, e2 := ioutil.ReadAll(inf)
		if e2 != nil { continue } // Couldn't read it after all.
		return unpacked, nil
	}

	err = os.NewError("No gzipped data found.")
	return
}

// Return places where the magic byte sequence appears in data.
func magicSites(data []byte, magic[]byte) []int {
	points := new(vector.IntVector)

        for i := 0; i < len(data) - len(magic); i++ {
		found := true
		for j := 0; j < len(magic); j++ {
			if data[i + j] != magic[j] { found = false }
		}
		if found { points.Push(i) }
	}

	return points.Data()
}
