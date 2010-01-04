// Package babble implements Bubble Babble encoding and decoding, as specified
// by http://wiki.yak.net/589.
package babble

import (
	"os"
	"strconv"
	"strings"
)

// The table of Babble vowels.
var vow = strings.Bytes("aeiouy")

// The table of Babble consonants.
var con = strings.Bytes("bcdfghklmnprstvzx")

// updateChecksum calculates a new Babble checksum value based on the next two
// bytes of input data.
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

type CorruptInputError int64

func (e CorruptInputError) String() string {
	return "illegal Bubble Babble data at input byte " + strconv.Itoa64(int64(e))
}

// devowel converts Babble vowels into the corresponding data values.
func devowel(char byte) (idx byte, ok bool) {
	for i, c := range vow {
		if char == c {
			return byte(i), true
		}
	}
	return 0, false
}

// deconsonant converts Babble consonants into the corresponding data values.
func deconsonant(char byte) (idx byte, ok bool) {
	for i, c := range con {
		if char == c {
			return byte(i), true
		}
	}
	return 0, false
}

// hyphen returns an error if the parameter character is not '-'. It has the
// same function signature as devowel and deconsonant so that it's func value
// can be used in the same type context as theirs.
func hyphen(char byte) (dummy byte, ok bool) { return 0, char == '-' }

// getTuple3 converts a sequence of vowel, consonant, vowel into three numeric
// values.
func getTuple3(offset int, src []byte) (result [3]byte, err os.Error) {
	lut := [](func(byte) (byte, bool)){devowel, deconsonant, devowel}
	for i := 0; i < 3; i++ {
		val, ok := lut[i](src[i])
		if !ok {
			err = CorruptInputError(offset + i)
			return
		}
		result[i] = val
	}
	return
}

// decode3WayByte decodes a byte that has been encoded into three Babble
// characters. Returns an error if the data is invalid or if it fails a
// checksum check.
func decode3WayByte(offset int, t [3]byte, c byte) (result byte, err os.Error) {
	high2 := (int(t[0]) - int(c%6) + 6) % 6
	if high2 >= 4 {
		err = CorruptInputError(offset)
		return
	}
	if t[1] > 16 {
		err = CorruptInputError(offset + 1)
		return
	}
	mid4 := int(t[1])
	low2 := (int(t[2]) - int(c/6%6) + 6) % 6
	if low2 >= 4 {
		err = CorruptInputError(offset + 2)
		return
	}
	result = byte(high2<<6) | byte(mid4<<2) | byte(low2)
	return
}

// verifyChecksumTuple checks that the checksum values are correct for a
// non-data-carrying terminating Babble tuple.
func verifyChecksumTuple(offset int, c byte, t [3]byte) os.Error {
	switch {
	case t[0] != c%6:
		return CorruptInputError(offset)
	case t[2] != c/6:
		return CorruptInputError(offset + 2)
	}
	return nil
}

// getByte3 decodes the part of the Babble string where three letters make up
// a byte and which can also terminate the Babble string, in which case isLast
// will be true. If the part is a terminating one, it might not carry byte
// data, in which case hasByte will be false.
func getByte3(offset int, src []byte, c byte) (result byte, isLast, hasByte bool, err os.Error) {
	// Must have at least one character beyond the three that are looked at
	// next. Either the start of the next tuple or the terminating 'x'.
	if len(src) < 4 {
		err = CorruptInputError(offset + len(src))
		return
	}

	// If the middle character is 'x', the last triple holds no byte payload,
	// and it's just checksum data instead.
	hasByte = src[1] != 'x'

	// A final 'x' terminates the data.
	isLast = src[3] == 'x'

	t, err := getTuple3(offset, src)
	if err != nil {
		return
	}

	if !hasByte {
		if !isLast {
			// Byteless checksum tuple not at the end of the data is an error.
			err = CorruptInputError(offset + 3)
			return
		}
		// Verify that the checksum is ok for byteless tuples.
		err = verifyChecksumTuple(offset, c, t)
		return
	}

	result, err = decode3WayByte(offset, t, c)
	return
}

// getTuple2 converts two consonants separated by a hyphen into two numerical
// values.
func getTuple2(offset int, src []byte) (result [2]byte, err os.Error) {
	lut := [](func(byte) (byte, bool)){deconsonant, hyphen, deconsonant}
	for i := 0; i < 3; i++ {
		val, ok := lut[i](src[i])
		if !ok {
			err = CorruptInputError(offset + i)
			return
		}
		switch i {
		case 0:
			result[0] = val
		case 2:
			result[1] = val
		}
	}
	return
}

// decode2WayByte decodes a byte that has been encoded into two Babble
// characters. This type of encoding uses all the available bits to represent
// data, so a checksum value is not used.
func decode2WayByte(offset int, t [2]byte) (result byte, err os.Error) {
	if t[0] > 16 {
		err = CorruptInputError(offset)
		return
	}
	if t[1] > 16 {
		err = CorruptInputError(offset + 1)
		return
	}

	result = (t[0] << 4) | t[1]
	return
}

// getByte2 decodes the part of the Babble string where two letters separated
// by a hyphen make up a byte. Doesn't use a checksum, since all letter bytes
// are taken up by the byte payload.
func getByte2(offset int, src []byte) (result byte, err os.Error) {
	// The second part, two-letter tuple with a hyphen in the middle.
	if len(src) < 3 {
		err = CorruptInputError(offset + len(src))
		return
	}

	t, err := getTuple2(offset, src)
	if err != nil {
		return
	}

	result, err = decode2WayByte(offset, t)

	return
}

// Decode decodes a Babble string into the corresponding byte array. Returns
// the number of bytes decoded, and an error if the string isn't a Babble
// string. Once Decode encounters a Babble data terminator in the src data, it
// stops decoding and returns the number of bytes read, regardless of whether
// there is more data remaining.
func Decode(dst, src []byte) (n int, err os.Error) {
	c := byte(1)

	if len(src) == 0 {
		err = CorruptInputError(0)
		return
	}

	// Babble strings must start with 'x'.
	if src[0] != 'x' {
		err = CorruptInputError(0)
		return
	}

	src = src[1:len(src)]

	offset := 1

	// Decode the full tuples.
	for {
		b1, wasLast, hadLastByte, err := getByte3(offset, src, c)
		if err != nil {
			return n, err
		}
		if wasLast {
			if hadLastByte {
				dst[n] = b1
				n++
			}
			return
		}

		dst[n] = b1
		n++
		src = src[3:len(src)]
		offset += 3

		b2, err := getByte2(offset, src)
		if err != nil {
			return n, err
		}

		dst[n] = b2
		n++
		src = src[3:len(src)]
		offset += 3

		c = updateChecksum(c, b1, b2)
	}

	return
}

// DecodeString decodes a babble string, returning the resulting byte array.
func DecodeString(src string) (result []byte, err os.Error) {
	result = make([]byte, MaxDecodedLen(len(src)))
	n, err := Decode(result, strings.Bytes(src))
	if err != nil {
		return
	}
	result = result[0:n]
	return
}
