// chunks.go
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
	"teratogen/space"
	"teratogen/world"
)

type chunkCache struct {
	index  map[string]int
	chunks []*Chunk
}

func (cs *chunkCache) Add(c *Chunk) {
	key := c.String()
	// Chunk is already stored.
	if _, ok := cs.index[key]; ok {
		return
	}

	cs.index[key] = len(cs.chunks)
	cs.chunks = append(cs.chunks, c)
}

func (cs *chunkCache) AddAll(c *Chunk) {
	set := []*Chunk{c}
	for i := 0; i < 3; i++ {
		set = append(set, set[len(set)-1].RotateCW())
	}
	for i := 0; i < 4; i++ {
		set = append(set, set[len(set)-4].MirrorX())
	}
	for _, chunk := range set {
		cs.Add(chunk)
	}
}

func buildCache(asciiMaps []string) *chunkCache {
	result := &chunkCache{make(map[string]int), []*Chunk{}}

	for _, asciiMap := range asciiMaps {
		chunk, err := ParseChunk(asciiMap)
		if err != nil {
			panic("Bad input for chunk cache ")
		}
		result.AddAll(chunk)
	}
	return result
}

var cache = buildCache(chunkData)

func Chunks() []*Chunk {
	return cache.chunks
}

type ChunkSet map[image.Point]*Chunk

func (c ChunkSet) Place(w *world.World, loc space.Location) {
	for p, chunk := range c {
		chunk.Place(w, w.Manifold.Offset(loc, p.Mul(ChunkSpan)))
	}
}
