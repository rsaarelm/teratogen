// chunkgraph.go
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
	"sort"
)

type ChunkGraph struct {
	edge        map[image.Point]bool
	dummyChunks map[image.Point]*Chunk
	chunks      map[image.Point]*Chunk
}

func NewChunkGraph() *ChunkGraph {
	return &ChunkGraph{
		edge:        make(map[image.Point]bool),
		dummyChunks: make(map[image.Point]*Chunk),
		chunks:      make(map[image.Point]*Chunk)}
}

func (c *ChunkGraph) ChunkAt(p image.Point) *Chunk {
	if result, ok := c.dummyChunks[p]; ok {
		return result
	}
	if result, ok := c.chunks[p]; ok {
		return result
	}
	return nil
}

func (c *ChunkGraph) OpenSlots() []image.Point {
	result := []image.Point{}
	for pt, _ := range c.edge {
		result = append(result, pt)
	}

	// To ensure that mapgen stays deterministic, compensate for the
	// unspecified map iteration order and sort the result before returning.
	sort.Sort(sortPoints(result))
	return result
}

type sortPoints []image.Point

func (p sortPoints) Len() int { return len(p) }

func (p sortPoints) Swap(i, j int) { p[i], p[j] = p[j], p[i] }

func (p sortPoints) Less(i, j int) bool {
	switch {
	case p[i].Y < p[j].Y:
		return true
	case p[i].Y == p[j].Y:
		return p[i].X < p[j].X
	}
	return false
}

func (c *ChunkGraph) FittingChunks(p image.Point, chunks []*Chunk) []*Chunk {
	result := []*Chunk{}
	for _, chunk := range chunks {
		if c.fits(p, chunk) {
			result = append(result, chunk)
		}
	}
	return result
}

func (c *ChunkGraph) PlaceChunk(p image.Point, chunk *Chunk) {
	if !c.fits(p, chunk) {
		panic("Adding a chunk that will not fit")
	}
	c.chunks[p] = chunk

	delete(c.edge, p)
	for dir, offset := range []image.Point{{0, -1}, {1, 0}, {0, 1}, {-1, 0}} {
		p2 := p.Add(offset)
		if c.ChunkAt(p2) == nil && chunk.Edge(EdgeDir(dir)).IsOpen() {
			c.edge[p2] = true
		}
	}
}

func (c *ChunkGraph) PlaceDummyChunk(p image.Point, chunk *Chunk) {
	if _, ok := c.chunks[p]; ok {
		panic("Placing dummy chunk over a real one")
	}
	c.dummyChunks[p] = chunk
}

func (c *ChunkGraph) Chunks() ChunkSet {
	return ChunkSet(c.chunks)
}

func (c *ChunkGraph) fits(p image.Point, chunk *Chunk) bool {
	if chunk == nil {
		panic("Fitting a nil chunk")
	}
	if c.ChunkAt(p) != nil {
		return false
	}
	for dir, offset := range []image.Point{{0, -1}, {1, 0}, {0, 1}, {-1, 0}} {
		p2, d := p.Add(offset), EdgeDir(dir)
		chunk2 := c.ChunkAt(p2)
		if chunk2 == nil {
			continue
		}
		// Edge mismatch
		if chunk.Edge(d) != chunk2.Edge(d.Opposite()) {
			return false
		}
	}
	return true
}
