// gen.go
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
)

type Gen struct {
	pegs  *pegMap
	cells map[image.Point]MapCell
	wall  MapCell
}

type OffsetChunk struct {
	chunk  *Chunk
	offset image.Point
}

func (oc OffsetChunk) Bounds() image.Rectangle {
	return image.Rectangle{image.Pt(0, 0), oc.chunk.Dim()}.Add(oc.offset)
}

func (oc OffsetChunk) Offset() image.Point {
	return oc.offset
}

func (oc OffsetChunk) At(pt image.Point) (MapCell, bool) {
	if cell, ok := oc.chunk.cells[pt.Sub(oc.offset)]; ok {
		return cell, true
	}
	return spaceCell, false
}

func (oc OffsetChunk) Pegs() []Peg {
	result := []Peg{}
	for _, peg := range oc.chunk.pegs {
		result = append(result, peg.Add(oc.offset))
	}
	return result
}

func New(initial *Chunk, wall MapCell) *Gen {
	result := &Gen{newPegMap(), map[image.Point]MapCell{}, wall}
	result.AddChunk(OffsetChunk{initial, initial.dim.Div(-2)})
	return result
}

func (cg *Gen) At(pt image.Point) (cell MapCell, ok bool) {
	cell, ok = cg.cells[pt]
	return
}

func (cg *Gen) PegsAt(pt image.Point) []Peg {
	return cg.pegs.At(pt)
}

func (cg *Gen) Map() map[image.Point]MapCell {
	return cg.cells
}

func (cg *Gen) OpenPegs() []Peg {
	return cg.pegs.pegs
}

func (cg *Gen) sealPeg(peg Peg) {
	for _, pt := range peg.Points() {
		cg.cells[pt] = cg.wall
	}
}

// ClosePeg removes the Peg from the chunkmap generator's set of open pegs and
func (cg *Gen) ClosePeg(peg Peg) {
	cg.pegs.Remove(peg)
	cg.sealPeg(peg)
}

func (cg *Gen) CloseAllPegs() {
	for _, peg := range cg.OpenPegs() {
		cg.ClosePeg(peg)
	}
}

// FittingChunks returns the set of offset-wrapped chunks that can be attached
// to the given Peg on the chunkmap generator.
func (cg *Gen) FittingChunks(peg Peg, chunks []*Chunk) []OffsetChunk {
	result := []OffsetChunk{}
	for _, ch := range chunks {
		for _, opposingPeg := range ch.pegs {
			// The same chunk can show up multiple times with different
			// offsets in the result if it has multiple fitting pegs.
			oc := OffsetChunk{ch, peg.offset.Sub(opposingPeg.offset)}
			if opposingPeg.latches(peg) && cg.fits(oc) {
				result = append(result, oc)
			}
		}
	}
	return result
}

func (cg *Gen) fits(oc OffsetChunk) bool {
	inside := oc.chunk.InsideBounds()
	for pt, cell := range oc.chunk.cells {
		if pt.In(inside) {
			if _, ok := cg.cells[pt.Add(oc.offset)]; ok {
				return false
			}
		} else {
			if existingCell, ok := cg.cells[pt.Add(oc.offset)]; ok {
				// Non-peg cells in a chunk can overwrite peg cells in the
				// existing map.
				overwritingPeg := oc.chunk.isPegCell(existingCell) && !oc.chunk.isPegCell(cell)
				if !overwritingPeg && cell != existingCell {
					return false
				}
			}
		}
	}
	return true
}

func (cg *Gen) AddChunk(oc OffsetChunk) {
	cg.closeOldPegs(oc)
	cg.addAndHandleNewPegs(oc)
}

func (cg *Gen) closeOldPegs(oc OffsetChunk) {
	futureChart := comboMappable([]mappable{cg, oc})

	bounds := oc.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			pt := image.Pt(x, y)
			if _, ok := oc.At(pt); !ok {
				continue
			}

			for _, peg := range cg.pegs.At(pt) {
				if !peg.matches(futureChart) {
					// An existing peg falls on the new chunk's area, and the
					// new chunk is about to introduce terrain that the
					// existing peg will not match. Wall out the existing peg.
					cg.sealPeg(peg)
				}
				if peg.covered(oc) {
					cg.pegs.Remove(peg)
				}
			}
		}
	}
}

func (cg *Gen) addAndHandleNewPegs(oc OffsetChunk) {
	brush := map[image.Point]MapCell{}
	bounds := oc.Bounds()
	newMap := pegMap{oc.Pegs()}

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			pt := image.Pt(x, y)
			cell, ok := oc.At(pt)
			if !ok {
				continue
			}
			brush[pt] = cell
		}
	}

	for pt, _ := range brush {
		for _, peg := range newMap.At(pt) {
			if !peg.matches(cg) {
				// Seal the peg and remove the peg's area from the new chunk
				// brush.
				cg.sealPeg(peg)
				for _, pt := range peg.Points() {
					delete(brush, pt)
				}
				newMap.Remove(peg)
			} else if peg.covered(cg) {
				// Remove, but don't seal, pegs that are valid but already
				// covered by existing terrain and not viable as expansion
				// points.
				newMap.Remove(peg)
			}
		}
	}

	for _, peg := range newMap.pegs {
		cg.pegs.Add(peg)
	}
	for pt, cell := range brush {
		cg.cells[pt] = cell
	}
}

type comboMappable []mappable

func (c comboMappable) At(pt image.Point) (MapCell, bool) {
	for _, m := range c {
		if ret, ok := m.At(pt); ok {
			return ret, ok
		}
	}
	return spaceCell, false
}

// If this becomes slow, optimize it by turning it into a struct with an added
// mapping from points to indices in the peg array.
type pegMap struct {
	pegs []Peg
}

func newPegMap() *pegMap {
	return &pegMap{[]Peg{}}
}

func (pm *pegMap) Add(peg Peg) {
	pm.pegs = append(pm.pegs, peg)
}

func (pm *pegMap) Remove(peg Peg) {
	for i, p := range pm.pegs {
		if p == peg {
			pm.pegs[i] = pm.pegs[len(pm.pegs)-1]
			pm.pegs = pm.pegs[:len(pm.pegs)-1]
			return
		}
	}
	panic("Removing peg not in pegMap")
}

func (pm *pegMap) At(pos image.Point) []Peg {
	result := []Peg{}
	for _, p := range pm.pegs {
		for _, pt := range p.Points() {
			if pt == pos {
				result = append(result, p)
			}
		}
	}
	return result
}
