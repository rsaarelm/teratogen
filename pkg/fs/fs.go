// Archive and filesystem utilities

package fs

import (
	"bytes"
	"compress/gzip"
	"container/vector"
//	"fmt"
	"io/ioutil"
	"os"
)

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
	// Don't write the magic byte as a straight literal, as that'll then
	// show up as an extra magic string in the binary code. The third
	// character is not an official part of the gz file magic sequence. It
	// indicates compression method deflate, which seems to be used in
	// most gz files. This is great for decreasing false positives, but it
	// may fail with exotic gz files.
	gzMagic := make([]byte, 3)
	gzMagic[0] = 0x1f
	gzMagic[1] = 0x8b
	gzMagic[2] = 0x08

	sites := magicSites(fileData, gzMagic)

	for n := range sites {
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
func magicSites(data []byte, magic[]byte) (result []int) {
	points := new(vector.Vector)

        for i := 0; i < len(data) - len(magic); i++ {
		found := true
		for j := 0; j < len(magic); j++ {
			if data[i + j] != magic[j] { found = false }
		}
		if found { points.Push(i) }
	}

	result = make([]int, points.Len())
	for i := 0; i < points.Len(); i++ {
		result[i] = points.At(i).(int)
	}
	return
}

func looksLikeGz(data []byte) bool {
	// Try to init the reader. Seems to only check for the magic number,
	// so not that much use since we already identify the candidates by
	// that.
	inf, e1 := gzip.NewInflater(bytes.NewBuffer(data))
	if e1 != nil { return false }
	// Try to read a few bytes. This is generally where random occurrences
	// of the magic id fail.
	_, e2 := inf.Read(make([]byte, 16))
	if e2 != nil { return false }
	return true
}
