/* babble_test.go

   Copyright (C) 2012 Risto Saarelma

   This program is free software: you can redistribute it and/or modify
   it under the terms of the GNU General Public License as published by
   the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.

   This program is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU General Public License for more details.

   You should have received a copy of the GNU General Public License
   along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

package babble

import (
	"bytes"
	"testing"
)

type testpair struct {
	decoded, encoded []byte
}

func makePair(dec, enc string) testpair { return testpair{[]byte(dec), []byte(enc)} }

var pairs = []testpair{
	makePair("", "xexax"),
	makePair("1234567890", "xesef-disof-gytuf-katof-movif-baxux"),
	makePair("Pineapple", "xigak-nyryk-humil-bosek-sonax"),
	makePair("asdf", "ximel-finek-koxex"),
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

func TestDecodeWithTrailingJunk(t *testing.T) {
	for i, p := range pairs {
		src := []byte(string(p.encoded) + "-trail-ingju-nkx")
		dst := make([]byte, MaxDecodedLen(len(src)))
		n, err := Decode(dst, src)
		dst = dst[0:n]
		if err != nil {
			t.Errorf("#%d: Decoding %#v caused error %#v", i, string(src), err)
		}
		if bytes.Compare(dst, p.decoded) != 0 {
			t.Errorf("#%d: Decode decoded %#v, expected %#v", i, string(dst), string(src))
		}
	}
}

func TestDecode(t *testing.T) {
	for i, p := range pairs {
		dst := make([]byte, MaxDecodedLen(len(p.encoded)))
		n, err := Decode(dst, p.encoded)
		dst = dst[0:n]
		if err != nil {
			t.Errorf("#%d: Decoding %#v caused error %#v", i, string(p.encoded), err)
		}
		if bytes.Compare(dst, p.decoded) != 0 {
			t.Errorf("#%d: Decode decoded %#v, expected %#v", i, string(dst), string(p.decoded))
		}
	}
}

type corruptPair struct {
	encoded []byte
	err     error
}

func makeCorrupt(enc string, err error) corruptPair {
	return corruptPair{[]byte(enc), err}
}

var corrupts = []corruptPair{
	makeCorrupt("Ph'nglui mglw'nafh Cthulhu R'lyeh wgah'nagl fhtagn", CorruptInputError(0)),
	makeCorrupt("xyxax", CorruptInputError(1)),
	makeCorrupt("xexux", CorruptInputError(3)),
	makeCorrupt("xexak", CorruptInputError(4)),
	makeCorrupt("nyryk-humil-bosek", CorruptInputError(0)),
	makeCorrupt("", CorruptInputError(0)),
	makeCorrupt("x", CorruptInputError(1)),
	makeCorrupt("xi", CorruptInputError(2)),
	makeCorrupt("xig", CorruptInputError(3)),
	makeCorrupt("xiga", CorruptInputError(4)),
	makeCorrupt("xigak", CorruptInputError(5)),
	makeCorrupt("xigak-", CorruptInputError(6)),
	makeCorrupt("xigak-n", CorruptInputError(7)),
	makeCorrupt("xigak-ny", CorruptInputError(8)),
	makeCorrupt("xigak-nyr", CorruptInputError(9)),
	makeCorrupt("xigak-nyry", CorruptInputError(10)),
	makeCorrupt("xigak-nyryk", CorruptInputError(11)),
	makeCorrupt("xigak-nyryk-", CorruptInputError(12)),
	makeCorrupt("xigak-nyryk-humil-bosek", CorruptInputError(23)),
	makeCorrupt("nyryk-humil-bosek-sonax", CorruptInputError(0)),
}

func TestDecodeCorrupt(t *testing.T) {
	for i, corrupt := range corrupts {
		dst := make([]byte, MaxDecodedLen(len(corrupt.encoded)))
		_, err := Decode(dst, corrupt.encoded)
		if err == nil {
			t.Errorf("#%d: Decoder failed to detect corruption in %#v", i, string(corrupt.encoded))
		}
		if err != corrupt.err {
			t.Errorf("#%d (%s): Expected err %#v, got %#v", i, string(corrupt.encoded), corrupt.err, err)
		}
	}
}
