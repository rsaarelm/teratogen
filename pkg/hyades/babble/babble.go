// This package implements Bubble Babble (http://wiki.yak.net/589) encoding and decoding

package babble

import (
	"fmt"
	"os"
	"strings"
)

var vow = strings.Bytes("aeiouy")
var con = strings.Bytes("bcdfghklmnprstvzx")

func updateChecksum(c, data1, data2 byte) byte {
	return byte((int(c)*5 + (int(data1)*7 + int(data2))) % 36)
}

func EncodedLen(n int) int {
	nTuples := n / 2
	partialTuple := 3
	terminators := 2
	hyphens := nTuples
	return 5*nTuples + hyphens + partialTuple + terminators
}

func MaxDecodedLen(n int) int {
	if n == 5 {
		// Only the partial tuple present.
		return 1
	}
	nTuples := (n + 1) / 6
	return nTuples * 2
}

// Encode encodes src into EncodedLen(len(src)) bytes of dst as Bubble Babble
// code.
func Encode(dst, src []byte) int {
	dst[0] = 'x'
	var c byte = 1
	numIter := len(src)/2 + 1
	fmt.Printf("Data: %#v, numiter: %d\n", string(src), numIter)

	for i := 0; i < numIter; i++ {
		if i+1 < numIter || len(src)%2 != 0 {
			d1 := src[i*2]

			dst[i*6+1] = vow[(((d1>>6)&3)+c)%6]
			dst[i*6+2] = con[(d1>>2)&15]
			dst[i*6+3] = vow[((d1&3)+c/6)%6]

			if i+1 < numIter {
				d2 := src[i*2+1]
				// Haven't written the last part yet.
				dst[i*6+4] = con[(d2>>4)&15]
				dst[i*6+5] = '-'
				dst[i*6+6] = con[(d2&15)%36]
				c = updateChecksum(c, d1, d2)
			}
		} else {
			// Last part for even data length.
			dst[i*6+1] = vow[c%6]
			dst[i*6+2] = con[16]
			dst[i*6+3] = vow[c/6]
		}
	}
	dst[(len(src)/2)*6+4] = 'x'
	return EncodedLen(len(src))
}

// EncodeToString returns the Bubble Babble encoding of src.
func EncodeToString(src []byte) string {
	dst := make([]byte, EncodedLen(len(src)))
	Encode(dst, src)
	return string(dst)
}

func Decode(dst, src []byte) (n int, err os.Error) {
	return
}
