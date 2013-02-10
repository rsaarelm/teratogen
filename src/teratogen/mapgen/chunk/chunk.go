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

package chunk

import (
	"errors"
	"image"
	"math"
	"strings"
	"unicode"
	"unicode/utf8"
)

type Chunk struct {
	pegs  []Peg
	cells map[image.Point]MapCell
	dim   image.Point
	spec  *ParseSpec
}

func (c *Chunk) Dim() image.Point {
	return c.dim
}

func (c *Chunk) Bounds() image.Rectangle {
	return image.Rectangle{image.Pt(0, 0), c.Dim()}
}

func (c *Chunk) InsideBounds() image.Rectangle {
	return c.Bounds().Inset(1)
}

func (c *Chunk) RotatedCW() *Chunk {
	str := ""
	for y := 0; y < c.dim.X; y++ {
		for x := 0; x < c.dim.Y; x++ {
			if cell, ok := c.cells[image.Pt(y, c.dim.Y-1-x)]; ok {
				str += string(cell)
			} else {
				str += " "
			}
		}
		str += "\n"
	}

	result, err := ParseChunk(*c.spec, str)
	if err != nil {
		panic("Parsing rotated chunk failed")
	}
	return result
}

func (c *Chunk) MirroredX() *Chunk {
	str := ""
	for y := 0; y < c.dim.Y; y++ {
		for x := 0; x < c.dim.X; x++ {
			if cell, ok := c.cells[image.Pt(c.dim.X-1-x, y)]; ok {
				str += string(cell)
			} else {
				str += " "
			}
		}
		str += "\n"
	}

	result, err := ParseChunk(*c.spec, str)
	if err != nil {
		panic("Parsing mirrored chunk failed")
	}
	return result
}

func (c *Chunk) String() string {
	str := ""
	for y := 0; y < c.dim.Y; y++ {
		for x := 0; x < c.dim.X; x++ {
			if cell, ok := c.cells[image.Pt(x, y)]; ok {
				str += string(cell)
			} else {
				str += " "
			}
		}
		str += "\n"
	}
	return str
}

func (c *Chunk) isPegCell(cell MapCell) bool {
	return strings.ContainsRune(c.spec.PegCells, rune(cell))
}

func generateChunkVariants(knownChunks map[string]bool, chunk *Chunk) []*Chunk {
	variants := []*Chunk{chunk}
	for i := 0; i < 3; i++ {
		variants = append(variants, variants[len(variants)-1].RotatedCW())
	}
	for i := 0; i < 4; i++ {
		variants = append(variants, variants[len(variants)-4].MirroredX())
	}

	result := []*Chunk{}
	for _, ch := range variants {
		str := ch.String()
		if _, ok := knownChunks[str]; !ok {
			knownChunks[str] = true
			result = append(result, ch)
		}
	}
	return result
}

// GenerateChunkVariants expands a given chunk list with all the unique
// mirrored and rotated versions of the chunks on the list.
func GenerateChunkVariants(chunks []*Chunk) []*Chunk {
	knownChunks := map[string]bool{}
	result := []*Chunk{}
	for _, c := range chunks {
		result = append(result, generateChunkVariants(knownChunks, c)...)
	}
	return result
}

func ParseChunk(spec ParseSpec, asciiMap string) (result *Chunk, err error) {
	var lines []string
	lines, err = preprocess(asciiMap)
	if err != nil {
		return
	}

	w, h := len(lines[0]), len(lines)
	result = &Chunk{[]Peg{}, map[image.Point]MapCell{}, image.Pt(w, h), &spec}
	edges := extractEdges(lines)
	for dir, edge := range edges {
		var pegs []Peg
		pegs, err = collectPegs(corner(Dir4(dir), result.dim), Dir4(dir), spec, edge)
		if err != nil {
			return
		}
		for _, peg := range pegs {
			result.pegs = append(result.pegs, peg)
		}
	}

	for y, line := range lines {
		for x, ch := range line {
			if !unicode.IsSpace(ch) {
				result.cells[image.Pt(x, y)] = MapCell(ch)
			}
		}
	}

	return
}

// preprocess extracts the rectangle of lines that contain map data from the input.
func preprocess(asciiMap string) (lines []string, err error) {
	lines = strings.Split(asciiMap, "\n")

	// Remove heading and trailing empty lines.
	for len(lines) > 0 && isEmpty(lines[0]) {
		lines = lines[1:len(lines)]
	}
	for len(lines) > 0 && isEmpty(lines[len(lines)-1]) {
		lines = lines[0 : len(lines)-1]
	}

	if len(lines) == 0 {
		err = errors.New("Empty ASCII")
		return
	}

	indent := math.MaxInt32
	maxWidth := 0

	// Deterimine indent level and rightmost text level, look for forbidden things
	for _, line := range lines {
		if strings.Contains(line, "\t") {
			err = errors.New("Physical tabs in chunk ASCII")
			return
		}
		if isEmpty(line) {
			err = errors.New("Discontinuous ASCII map")
			return
		}
		startPos, endPos := measure(line)
		if startPos >= 0 {
			if startPos < indent {
				indent = startPos
			}
			if endPos > maxWidth {
				maxWidth = endPos
			}
		}
	}

	// Correct the width for the indent level.
	maxWidth -= indent

	// Pad all lines to the same width.
	for i, line := range lines {
		line = strings.TrimRightFunc(line, unicode.IsSpace)
		line = line[indent:len(line)]
		pad := maxWidth - len(line)
		if pad > 0 {
			line += strings.Repeat(" ", pad)
		}
		if len(line) != maxWidth {
			panic("Line padding broken")
		}
		lines[i] = line
	}
	return
}

func isEmpty(line string) bool {
	return len(strings.TrimSpace(line)) == 0
}

func measure(line string) (startPos int, endPos int) {
	startPos = len(line) - len(strings.TrimLeftFunc(line, unicode.IsSpace))
	endPos = len(strings.TrimRightFunc(line, unicode.IsSpace))
	return
}

func extractEdges(block []string) (result [4]string) {
	result[north] = block[0]
	result[south] = block[len(block)-1]
	for _, line := range block {
		result[west] += string(line[0])
		result[east] += string(line[len(line)-1])
	}
	return
}

func corner(facing Dir4, dim image.Point) image.Point {
	switch facing {
	case north:
		return image.Pt(0, 0)
	case east:
		return image.Pt(dim.X-1, 0)
	case south:
		return image.Pt(0, dim.Y-1)
	case west:
		return image.Pt(0, 0)
	}
	panic("Bad facing")
}

func collectPegs(origin image.Point, facing Dir4, spec ParseSpec, edge string) (result []Peg, err error) {
	result = []Peg{}
	inPeg := false
	for x, ch := range edge {
		if spec.isPeg(uint8(ch)) {
			if !inPeg {
				inPeg = true
				start := origin.Add(facing.alongVec().Mul(x))
				result = append(result, Peg{start, facing, ""})
			}
			current := &result[len(result)-1]
			current.pattern += string(ch)
		} else if spec.isOverlap(uint8(ch)) {
			if x == 0 || x == len(edge)-1 || !spec.isPeg(edge[x-1]) ||
				!spec.isPeg(edge[x+1]) {
				err = errors.New("Overlap char not between two peg chars")
				return
			}
			if !inPeg {
				panic("Not in peg even though previous char is peg char")
			}

			// The overlap char means that the current and the next peg share
			// the current cell. End the current peg by repeating its last
			// character.
			current := &result[len(result)-1]
			last, _ := utf8.DecodeLastRuneInString(current.pattern)
			current.pattern += string(last)

			// Start the next peg with the same character as its second one.
			start := origin.Add(facing.alongVec().Mul(x))
			result = append(result, Peg{start, facing, string(MapCell(edge[x+1]))})
		} else {
			inPeg = false
		}
	}

	return
}
