// chunkdata.go
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
	"teratogen/mapgen/chunk"
	"teratogen/world"
)

var legend = map[rune]placeFn{
	'#': terrainPlacer(world.WallTerrain),
	'.': terrainPlacer(world.FloorTerrain),
	'|': terrainPlacer(world.DoorTerrain),
	'b': terrainPlacer(world.BarrelTerrain),
	'c': terrainPlacer(world.ChairTerrain),
	't': terrainPlacer(world.CounterTerrain),
	'p': terrainPlacer(world.PlantTerrain),
}

var chunkData = parseChunks(`
##|##
#...#
|...|
#...#
##|##

.....
.....
.....
.....
.....

##|##
.....
.....
.....
.....

#####
.....
.....
.....
.....

#####
#....
#....
#....
#....

.....
.....
.....
.....
#....


#####
#....
|....
#....
#....

#####
#.b..
#b.b.
#.b..
#....

##|###|##
#b.....b#
|.......|
#.......#
#.......#
#.p.....#
#ct.....|
#.p.....#
######|##

##|###|##
#.......#
|.......|
#.......#
#.......#
#.......#
#.......|
#.......#
##|###|##

##|###|###|##
#...........#
|...........|
#...........#
#...........#
#...........#
|...........|
#...........#
######|###|##

##|###|###|##
#...........#
|..bb.......|
#..bb.......#
#...........#
#...........#
|...........|
#...........#
##|###|###|##

##|###|##
#.......#
|.......|
#.......#
#...#####
#...#    
|...#    
#...#    
##|##    
`)

func parseChunks(chunkData string) []*chunk.Chunk {
	result := []*chunk.Chunk{}
	for _, asciiMap := range chunk.SplitMaps(chunkData) {
		chunk, err := chunk.Parse(chunk.ParseSpec{"|.", '*'}, asciiMap)
		if err != nil {
			panic(err)
		}
		result = append(result, chunk)
	}
	result = chunk.GenerateVariants(result)
	return result
}
