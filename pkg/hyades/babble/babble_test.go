package babble

import (
	"bytes"
	"strings"
	"testing"
)

type testpair struct {
	decoded, encoded []byte
}

var pairs = []testpair{
	testpair{strings.Bytes(""), strings.Bytes("xexax")},
	testpair{strings.Bytes("1234567890"), strings.Bytes("xesef-disof-gytuf-katof-movif-baxux")},
	testpair{strings.Bytes("Pineapple"), strings.Bytes("xigak-nyryk-humil-bosek-sonax")},
}

func TestEncode(t *testing.T) {
	for i, p := range pairs {
		dst := make([]byte, EncodedLen(len(p.decoded)))
		n := Encode(dst, p.decoded)
		if n != len(dst) {
			t.Errorf("#%d: Encode returned %d, expected %d", i, n, len(dst))
		}
		if bytes.Compare(dst, p.encoded) != 0 {
			t.Errorf("#%d: Encode encoded %#v, expected %#v", i, string(dst), string(p.encoded))
		}
	}
}

func TestDecode(t *testing.T) {
	for i, p := range pairs {
		dst := make([]byte, DecodedLen(len(p.encoded)))
		n, err := Decode(dst, p.encoded)
		if err != nil {
			t.Errorf("#%d: Decoding %#v caused error %#v", p.encoded, err)
		}
		if n != len(dst) {
			t.Errorf("#%d: Decode returned %d, expected %d", i, n, len(dst))
		}
		if bytes.Compare(dst, p.decoded) != 0 {
			t.Errorf("#%d: Decode decoded %#v, expected %#v", i, string(dst), string(p.decoded))
		}
	}
}

func TestDecodeCorrupt(t *testing.T) {
	corrupts := [][]byte{
		strings.Bytes("Ph'nglui mglw'nafh Cthulhu R'lyeh wgah'nagl fhtagn"),
		strings.Bytes(""),
		strings.Bytes("nyryk-humil-bosek"),
		strings.Bytes("xigak-nyryk-humil-bosek"),
		strings.Bytes("nyryk-humil-bosek-sonax"),
	}
	for _, corrupt := range corrupts {
		dst := make([]byte, DecodedLen(len(corrupt)))
		_, err := Decode(dst, corrupt)
		if err == nil {
			t.Errorf("Decoder failed to detect corruption in %#v", string(corrupt))
		}
	}
}
