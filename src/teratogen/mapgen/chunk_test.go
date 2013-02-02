// chunk_test.go
//
// Copyright (C) 2013 Risto Saarelma
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
	"image"
	"testing"
)

func TestChunk(t *testing.T) {
	chunks := buildCache([]string{
		`
#####
#####
#####
#####
#####`,
		`
.....
.....
.....
.....
.....`,
		`
##.##
#...#
.....
#...#
##.##`,
		`
#####
#####
#####
##.##
##.##`,
	}).chunks

	solid := makeEdge("#####")
	hole := makeEdge("##.##")
	open := makeEdge(".....")

	for dir, opposite := range []EdgeDir{South, West, North, East} {
		if EdgeDir(dir).Opposite() != opposite {
			t.Error("Dir", dir, "has incorrect opposite")
		}
	}

	for i := 0; i < 4; i++ {
		if chunks[0].Edge(EdgeDir(i)) != solid {
			t.Error("Solid chunk doesn't get solid edge")
		}
		if chunks[1].Edge(EdgeDir(i)) != open {
			t.Error("Open chunk doesn't get open edge")
		}
	}

	if chunks[3].Edge(South) != hole {
		t.Error("Tunnel edge not recognized")
	}

	cg := NewChunkGraph()
	cg.PlaceChunk(image.Pt(0, 0), chunks[2])
	for i := 0; i < 4; i++ {
		p := EdgeDir(i).Vec()
		for j := 0; j < 4; j++ {
			chunk := chunks[3+j]

			if cg.fits(p, chunk) != (i == j) {
				t.Error("Incorrect chunk fit", i, j)
			}
		}
	}
}
