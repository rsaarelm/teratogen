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

package chunk

import (
	"image"
	"testing"
)

func parseChunk(t *testing.T, asciiMap string) *Chunk {
	result, err := Parse(ParseSpec{"|.", '*'}, asciiMap)
	if err != nil {
		t.Fatal("Parsing chunk failed:", err)
	}
	if result == nil {
		t.Fatal("No chunk generated")
	}
	return result
}

func parseChunks(t *testing.T, asciiMaps string) []*Chunk {
	result := []*Chunk{}
	for _, asciiMap := range SplitMaps(asciiMaps) {
		result = append(result, parseChunk(t, asciiMap))
	}
	return result
}

func TestChunk(t *testing.T) {
	chunk := parseChunk(t, `
  ###.|.#
  #.....#
  #.....###
  |.......|
  |.......| 
  |.......*
  #........
  #........
  ###.|.###
`)
	if chunk.Dim() != image.Pt(9, 9) {
		t.Error("Bad chunk size: ", chunk.Dim())
	}

	pegOffsets := map[image.Point]Peg{}
	for _, p := range chunk.pegs {
		pegOffsets[p.offset] = p
	}

	if len(pegOffsets) != 5 {
		t.Error("Bad peg count")
	}

	for _, pt := range []image.Point{{3, 0}, {0, 3}, {3, 8}, {8, 3}, {8, 5}} {
		if _, ok := pegOffsets[pt]; !ok {
			t.Error("Peg at " + pt.String() + " not found")
		}
	}
}

func TestGen(t *testing.T) {
	chunk := parseChunk(t, `
  ###.|.#
  #.....#
  #.....###
  |.......|
  |.......|
  |.......*
  #........
  #........
  ###.|.###
`)

	gen := New(chunk, '#')

	if len(gen.OpenPegs()) != 5 {
		t.Errorf("Bad number of open pegs in generator: %d", len(gen.OpenPegs()))
	}
}

func TestCorner(t *testing.T) {
	chunks := parseChunks(t, `
...
...
...

...
...
###

..#
..#
###
`)

	gen := New(chunks[0], '#')

	eastPt, southPt := image.Pt(2, 1), image.Pt(1, 2)

	if len(gen.PegsAt(eastPt)) != 1 {
		t.Fatal("Expected peg not found")
	}
	if len(gen.PegsAt(southPt)) != 1 {
		t.Fatal("Expected peg not found")
	}

	// Add a chunk to south peg
	c := gen.FittingChunks(gen.PegsAt(southPt)[0], chunks)[1]
	gen.AddChunk(c)
	if len(gen.PegsAt(southPt)) != 0 {
		t.Error("Peg not closed after chunk added to it")
	}
	if len(gen.PegsAt(eastPt)) != 1 {
		t.Error("Adjacent peg closed when chunk added")
	}
}

func TestSideOverwrite(t *testing.T) {
	chunks := parseChunks(t, `
###
#..
#.#
#.|
###

###
..#
#|#

#|#
#.#
###

###
|.#
###
`)

	// Build an L-shaped room from the first two chunks.
	gen := New(chunks[0], '#')
	gen.AddChunk(gen.FittingChunks(gen.PegsAt(image.Pt(2, 1))[0], chunks)[0])

	// Add a chunk to south peg. The chunk that matches the peg but overwrites
	// the other door should be returned.
	fits := gen.FittingChunks(gen.PegsAt(image.Pt(2, 3))[0], chunks)
	if len(fits) != 1 {
		t.Error("Fitting chunks not found")
	}
}
