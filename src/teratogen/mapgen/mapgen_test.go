// mapgen_test.go
//
// Copyright (C) 2012 Risto Saarelma
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package mapgen

import (
	"bytes"
	"image"
	"testing"
)

const width = 32
const height = 32

type testMap [width * height]byte

var mapArea = image.Rect(0, 0, width, height)

func (tm *testMap) At(p image.Point) Terrain {
	if p.In(mapArea) {
		switch tm[p.X+width*p.Y] {
		case '.':
			return Open
		case '%':
			return Doorway
		default:
			return Solid
		}
	}
	return Solid
}

func (tm *testMap) Set(p image.Point, t Terrain) {
	if p.In(mapArea) {
		switch t {
		case Solid:
			tm[p.X+width*p.Y] = '#'
		case Open:
			tm[p.X+width*p.Y] = '.'
		case Doorway:
			tm[p.X+width*p.Y] = '%'
		}
	}
}

func (t *testMap) String() string {
	var b bytes.Buffer
	b.WriteByte('\n') // Heading newline makes it look better in tester log.
	for y := 0; y < height; y++ {
		b.Write(t[y*width : y*width+width])
		b.WriteByte('\n')
	}
	return b.String()
}

func TestBsp(t *testing.T) {
	buf := new(testMap)
	for i, _ := range buf {
		buf[i] = '#'
	}
	BspRooms(buf, mapArea)
	t.Log(buf)
}
