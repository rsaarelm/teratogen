// This package implements Bubble Babble (http://wiki.yak.net/589) encoding
// and decoding. Bubble Babble encodes arbitrary binary data into
// human-pronouncable pseudo-words, and helps humans learn to recognize short
// binary sequences on sight.

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

// EncodeLen returns the number of bytes an encoded n bytes will take.
func EncodedLen(n int) int {
	nTuples := n / 2
	partialTuple := 3
	terminators := 2
	hyphens := nTuples
	return 5*nTuples + hyphens + partialTuple + terminators
}

// MaxDecodedLen returns the maximum number of bytes a decoding of a Babble
// string of length n will take. There may be a difference of one byte in the
// result length for the same input length depending on the content.
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
	c := byte(1)
	numIter := len(src)/2 + 1

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

func devowel(char byte) (idx byte, err os.Error) {
	for i, c := range vow {
		if char == c {
			return byte(i), nil
		}
	}
	err = os.NewError(fmt.Sprintf("Expected babble vowel, got '%c'", char))
	return
}

func deconsonant(char byte) (idx byte, err os.Error) {
	for i, c := range con {
		if char == c {
			return byte(i), nil
		}
	}
	err = os.NewError(fmt.Sprintf("Expected babble consonant, got '%c'", char))
	return
}

func hyphen(char byte) (dummy byte, err os.Error) {
	if char != '-' {
		err = os.NewError(fmt.Sprintf("Expected '-', got '%c'", char))
	}
	return
}

// Converts a Bubble Babble string tuple into the corresponding byte tuple.
func decodeTuple(src []byte, decodeFullTuple bool) (result [5]byte, err os.Error) {
	lut := [](func(byte) (byte, os.Error)){devowel, deconsonant, devowel, deconsonant, hyphen, deconsonant}
	idx := []int{0, 1, 2, 3, -1, 4}
	for i := 0; i < 6; i++ {
		val, err := lut[i](src[i])
		if err != nil {
			return
		}
		if idx[i] >= 0 {
			result[idx[i]] = val
		}
		if i == 2 && !decodeFullTuple {
			return
		}
	}
	return
}

// Decode byte encoded by the first type of Babble encoding into three bytes.
func decode3WayByte(a1, a2, a3 byte, c byte) (result byte, err os.Error) {
	high2 := (int(a1) - int(c%6) + 6) % 6
	if high2 >= 4 {
		err = os.NewError("Checksum error")
		return
	}
	if a2 > 16 {
		err = os.NewError("Algorithm error: Illegal high bits in data.")
		return
	}
	mid4 := int(a2)
	low2 := (int(a3) - int(c/6%6) + 6) % 6
	if low2 >= 4 {
		err = os.NewError("Checksum error")
		return
	}
	result = byte(high2<<6) | byte(mid4<<2) | byte(low2)
	return
}

// Decode byte encoded by the second type of Babble encoding into two bytes.
// Doesn't use the checksum value.
func decode2WayByte(a1, a2 byte) (result byte, err os.Error) {
	if a1 > 16 || a2 > 16 {
		err = os.NewError("Algorithm error: Illegal high bits in data.")
		return
	}
	result = (a1 << 4) | a2
	return
}

// Decode decodes a Babble string into the corresponding byte array. Returns
// the number of bytes decoded, and an error if the string isn't a Babble string.
func Decode(dst, src []byte) (n int, err os.Error) {
	nTuples := len(src) / 6
	c := byte(1)

	if len(src) < 5 {
		err = os.NewError(fmt.Sprintf("Babble string %#v too short.", string(src)))
		return
	}
	if src[0] != 'x' {
		err = os.NewError("Invalid Babble string prefix.")
		return
	}
	if src[len(src)-1] != 'x' {
		err = os.NewError("Invalid Babble string suffix.")
		return
	}

	src = src[1:len(src)]
	for i := 0; i < nTuples; i++ {
		t, err := decodeTuple(src, true)
		if err != nil {
			return
		}

		src = src[6:len(src)]

		d1, err := decode3WayByte(t[0], t[1], t[2], c)
		if err != nil {
			return
		}

		d2, err := decode2WayByte(t[3], t[4])
		if err != nil {
			return
		}
		c = updateChecksum(c, d1, d2)
		dst[i*2] = d1
		dst[i*2+1] = d2
	}

	t, err := decodeTuple(src, false)
	if err != nil {
		return
	}

	if t[1] == 16 {
		n = nTuples * 2
		// No last byte.
		if t[0] != c%6 || t[2] != c/6 {
			err = os.NewError("Checksum error in final tuple")
			return
		}
	} else {
		n = nTuples*2 + 1
		// Decode last byte.

		d, err := decode3WayByte(t[0], t[1], t[2], c)
		if err != nil {
			return
		}
		dst[nTuples*2] = d
	}

	return
}

// DecodeString tries to decode a babble string. It returns a byte array if
// the decoding was successful and an error otherwise.
func DecodeString(src string) (result []byte, err os.Error) {
	result = make([]byte, MaxDecodedLen(len(src)))
	n, err := Decode(result, strings.Bytes(src))
	if err != nil {
		return
	}
	result = result[0:n]
	return
}
