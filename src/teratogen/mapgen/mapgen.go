// mapgen.go
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

// Package mapgen defines the system for generating new game levels and
// populating them with entities.
package mapgen

import (
	"errors"
	"image"
	"math/rand"
	"teratogen/entity"
	"teratogen/mapgen/chunk"
	"teratogen/space"
	"teratogen/world"
)

type Mapgen struct {
	world   *world.World
	chart   space.Chart
	openSet map[space.Location]bool
}

func New(w *world.World) *Mapgen {
	return &Mapgen{world: w}
}

func (m *Mapgen) TestMap(start space.Location, depth int) (entry, exit space.Location) {
	legend := map[rune]placeFn{
		'#': terrainPlacer(world.WallTerrain),
		'.': terrainPlacer(world.FloorTerrain),
		'|': terrainPlacer(world.DoorTerrain),
		'b': terrainPlacer(world.BarrelTerrain),
		'c': terrainPlacer(world.ChairTerrain),
		't': terrainPlacer(world.CounterTerrain),
		'p': terrainPlacer(world.PlantTerrain),

		// Downstairs cell gets special handling at mapgen, failing that, it gets
		// turned into floor.
		'>': terrainPlacer(world.FloorTerrain),
		'<': terrainPlacer(world.StairTerrain),
	}

	entrance := parseChunks(`
####|####
#.......#
#.......#
#..###..#
|..#<...|
#..###..#
#.......#
#.......#
####|####
`)

	exits := parseChunks(`
####|####
#.......#
#.......#
#..###..#
|...>#..|
#..###..#
#.......#
#.......#
####|####
`)

	chunks := parseChunks(`
####|####
#.......#
#.......#
#.......#
|.......|
#.......#
#.......#
#.......#
####|####

####|####
#.......#
#.bb....#
#.bb....#
|.......|
#.......#
#.......#
#.......#
####|####

####|####
#.......#
#....#..#
#..c.|..#
|.ptp#..|
######..#
#bb.....#
#bb....b#
####|####

####|####
#..p.p..#
#.......#
#p.ctc.p#
|..ctc..|
#p.ctc.p#
#.......#
#..p.p..#
####|####
`)
	chunks = chunk.GenerateVariants(chunks)

	m.chart = simpleChart(start)

	cg := chunk.New(entrance[rand.Intn(len(entrance))], '#')
	cg.SetGrid(image.Pt(4, 4))

	nRooms := 5 + depth/2
	for i := 0; i < nRooms; i++ {
		pegs := cg.OpenPegs()
		if len(pegs) == 0 {
			panic("Map ran out of expansion room")
		}

		peg := pegs[rand.Intn(len(pegs))]

		var placeChunks []chunk.OffsetChunk
		if i == nRooms-1 {
			placeChunks = cg.FittingChunks(peg, exits)
			// XXX: Hardcoded exit point. All exists are assumed to have the
			// exit tile at the same offset.
			exit = m.chart.At(placeChunks[0].Offset().Add(image.Pt(4, 4)))
		} else {
			placeChunks = cg.FittingChunks(peg, chunks)
		}
		if len(placeChunks) == 0 {
			panic("Can't expand map")
		}

		chunk := placeChunks[rand.Intn(len(placeChunks))]

		cg.AddChunk(chunk)
	}
	cg.CloseAllPegs()

	for pt, cell := range cg.Map() {
		fn, ok := legend[rune(cell)]
		if ok {
			fn(m.world, m.chart.At(pt))
		} else {
			panic("Unknown terrain type " + string(cell))
		}
	}

	entry = m.chart.At(image.Pt(4, 4))
	return
}

func findExit(oc chunk.OffsetChunk) (exit image.Point, ok bool) {
	for offset, cell := range oc.Chunk().Map() {
		if cell == chunk.MapCell('>') {
			ok = true
			exit = offset.Add(oc.Offset())
			return
		}
	}
	return
}

func (m *Mapgen) init(start space.Location) {
	m.chart = simpleChart(start)
	m.openSet = map[space.Location]bool{}
}

func (m *Mapgen) setOpen(loc space.Location, isOpen bool) {
	if isOpen {
		m.openSet[loc] = true
	} else {
		delete(m.openSet, loc)
	}
}

func (m *Mapgen) randomLoc() (loc space.Location) {
	// XXX: O(n) time.
	n := rand.Intn(len(m.openSet))
	for k, _ := range m.openSet {
		loc = k
		n--
		if n < 0 {
			break
		}
	}
	return
}

func (m *Mapgen) spawn(obj entity.Entity, loc space.Location) error {
	if !m.world.Fits(obj, loc) {
		return errors.New("Spawn won't fit")
	}

	m.world.Place(obj, loc)

	// Remove the points from the open set if the entity blocks movement.
	if b, ok := obj.(entity.BlockMove); ok && b.BlocksMove() {
		for _, footLoc := range m.world.Manifold.FootprintFor(obj, loc) {
			m.setOpen(footLoc, false)
		}
	}
	return nil
}

func (m *Mapgen) terrain(pt image.Point) world.TerrainData {
	return m.world.Terrain(m.chart.At(pt))
}

func (m *Mapgen) setTerrain(pt image.Point, t world.Terrain) {
	m.world.SetTerrain(m.chart.At(pt), t)
}

func (m *Mapgen) checkSurroundings(loc space.Location, mustBeOpen, mustBeClosed []image.Point) bool {
	for _, offset := range mustBeOpen {
		if _, ok := m.openSet[m.world.Manifold.Offset(loc, offset)]; !ok {
			return false
		}
	}

	for _, offset := range mustBeClosed {
		if _, ok := m.openSet[m.world.Manifold.Offset(loc, offset)]; ok {
			return false
		}
	}

	return true
}

func (m *Mapgen) isExitEnclosure(open space.Location) bool {
	// Look for a suitable enclosure site in the positive x direction from the
	// open location.

	// Don't care if the actual site is open or blocked, it will be replaced
	// with a portal to the next level anyway.
	return m.checkSurroundings(open,
		[]image.Point{{0, 0}},
		[]image.Point{{1, -1}, {2, 0}, {2, 1}, {1, 1}})
}

func (m *Mapgen) isEntryEnclosure(open space.Location) bool {
	return m.checkSurroundings(open,
		[]image.Point{{0, 0}},
		[]image.Point{{-2, -1}, {-2, 0}, {-1, 1}, {-1, -1}})
}

// simpleChart is a chart that pays no attention to portals in the manifold.
type simpleChart space.Location

func (s simpleChart) At(pt image.Point) space.Location {
	return space.Location(s).Add(pt)
}

type placeFn func(*world.World, space.Location)

func terrainPlacer(ter world.Terrain) placeFn {
	return func(w *world.World, loc space.Location) {
		w.SetTerrain(loc, ter)
	}
}

func parseChunks(chunkData string) []*chunk.Chunk {
	result := []*chunk.Chunk{}
	for _, asciiMap := range chunk.SplitMaps(chunkData) {
		chunk, err := chunk.Parse(chunk.ParseSpec{"|.", '*'}, asciiMap)
		if err != nil {
			panic(err)
		}
		result = append(result, chunk)
	}
	return result
}
