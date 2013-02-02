// chunk.go
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
	"errors"
	"image"
	"strings"
	"teratogen/space"
	"teratogen/world"
)

const ChunkSpan = 4

type Chunk struct {
	str     string
	placers map[image.Point]placeFn
	edges   [4]Edge
}

// ParseChunk turns an ASCII map in a string into a chunk. The map must be a
// square of ChunkSpan + 1 times ChunkSpan + 1 characters that the chunk
// system can translate into map generation instructions.
func ParseChunk(text string) (result *Chunk, err error) {
	var lines []string
	lines, err = preprocess(text)
	if err != nil {
		return
	}

	result = new(Chunk)
	result.placers = make(map[image.Point]placeFn)
	result.str = strings.Join(lines, "\n")

	for y, line := range lines {
		for x, ch := range line {
			// Set the placer.
			result.placers[image.Pt(-ChunkSpan/2+x, -ChunkSpan/2+y)] = legend[ch]

			// Build edge keys for the edges (x, y) falls on.

			// Byte offsets in the four edges to apply current cell to. -1 if not applicable.
			edgeVector := []int{-1, -1, -1, -1}
			if y == 0 { // North
				edgeVector[0] = x
			}
			if x == ChunkSpan { // East
				edgeVector[1] = y
			}
			if y == ChunkSpan { // South
				edgeVector[2] = x
			}
			if x == 0 { // West
				edgeVector[3] = y
			}
			for edge, idx := range edgeVector {
				if idx == -1 {
					continue
				}
				result.edges[edge][idx] = uint8(ch)
			}
		}
	}
	return
}

func preprocess(text string) (lines []string, err error) {
	text = strings.Trim(text, " \t\n")
	lines = strings.Split(text, "\n")
	if len(lines) != ChunkSpan+1 {
		err = errors.New("Ascii map has wrong height")
		return
	}
	for i, line := range lines {
		lines[i] = strings.Trim(line, " \t")
		if len(lines[i]) != ChunkSpan+1 {
			err = errors.New("Ascii map has wrong width")
			return
		}

		for _, ch := range lines[i] {
			if _, ok := legend[ch]; !ok {
				err = errors.New("Unknown glyph '" + string(ch) + "' in ascii map")
				return
			}
		}
	}
	return
}

func (c *Chunk) String() string {
	return c.str
}

func (c *Chunk) Place(w *world.World, loc space.Location) {
	for offset, placer := range c.placers {
		placeLoc := w.Manifold.Offset(loc, offset)
		placer(w, placeLoc)
	}
}

func (c *Chunk) Edge(dir EdgeDir) Edge {
	return c.edges[dir]
}

// CharAt returns the char at the Chunk's ASCII map representation at a given
// point.
func (c *Chunk) CharAt(p image.Point) uint8 {
	lines := strings.Split(c.String(), "\n")
	return lines[p.Y][p.X]
}

// RotateCW returns a copy of a Chunk rotate 90 degrees clockwise.
func (c *Chunk) RotateCW() *Chunk {
	text := ""
	for y := 0; y < ChunkSpan+1; y++ {
		for x := 0; x < ChunkSpan+1; x++ {
			text += string(c.CharAt(image.Pt(y, ChunkSpan-x)))
		}
		text += "\n"
	}
	result, err := ParseChunk(text)
	if err != nil {
		panic("RotateCW failed")
	}
	return result
}

// MirrorX returns a copy of a Chunk mirrored along its X axis.
func (c *Chunk) MirrorX() *Chunk {
	text := ""
	for y := 0; y < ChunkSpan+1; y++ {
		for x := 0; x < ChunkSpan+1; x++ {
			text += string(c.CharAt(image.Pt(ChunkSpan-x, y)))
		}
		text += "\n"
	}
	result, err := ParseChunk(text)
	if err != nil {
		panic("MirrorX failed")
	}
	return result
}

type placeFn func(*world.World, space.Location)

func terrainPlacer(ter world.Terrain) placeFn {
	return func(w *world.World, loc space.Location) {
		w.SetTerrain(loc, ter)
	}
}

// Edge is a value attached to one of the four edges of a mapgen chunk and
// must be equal to the value on the matching edge of the neighboring chunk
// for that edge.
type Edge [8]uint8

func makeEdge(pattern string) (result Edge) {
	for i, ch := range pattern {
		result[i] = uint8(ch)
	}
	return result
}

func (e Edge) String() string {
	result := ""
	for i := 0; i < ChunkSpan+1; i++ {
		result += string(e[i])
	}
	return result
}

func (e Edge) IsOpen() bool {
	for _, ch := range e {
		switch ch {
		case '.', '|':
			return true
		}
	}
	return false
}

type EdgeDir uint8

const (
	North EdgeDir = iota
	East
	South
	West
)

func (e EdgeDir) Opposite() EdgeDir {
	return (e + 2) % 4
}

func (e EdgeDir) Vec() image.Point {
	return []image.Point{{0, -1}, {1, 0}, {0, 1}, {-1, 0}}[e]
}
