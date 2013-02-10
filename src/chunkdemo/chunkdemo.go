// chunkdemo.go
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

package main

import (
	"github.com/nsf/termbox-go"
	"image"
	"math"
	"math/rand"
	"strings"
	"teratogen/mapgen/chunk"
)

const chunkFile = `
.....
.....
.....
.....
.....

#....
#....
##...
##...
#....

#####
###..
##...
#....
#....

#####
##.##
##..#
#...#
#...#

.....
.....
.....
#....
##...

#....
#....
.....
#....
#....

#####
#####
.....
#####
#####

##.##
##.##
.....
#####
#####

##.##
##.##
.....
##.##
##.##

##.##
##.##
...##
#####
#####

#####
#####
|....
#####
#####

#####
#####
|..##
#####
#####

##|##
#...#
|...#
#...#
#####

#########
#...f...#
#..f.f..#
#...f...#
#.......#
#.......#
|.......#
#.......#
#########

#########
#........
.........
#........
#.......*
#........
#~~~.....
#~~~~....
######+##
`

var chunks = []*chunk.Chunk{}

func draw(gen *chunk.Gen, pegIdx int, chunkIdx int) {
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
	min := image.Pt(math.MaxInt32, math.MaxInt32)
	for pt, _ := range gen.Map() {
		if pt.X < min.X {
			min.X = pt.X
		}
		if pt.Y < min.Y {
			min.Y = pt.Y
		}
	}

	min = min.Sub(image.Pt(4, 4))

	for pt, cell := range gen.Map() {
		nPegs := len(gen.PegsAt(pt))
		foreground := termbox.ColorWhite
		if len(gen.OpenPegs()) == 0 {
			foreground = termbox.ColorYellow
		}
		background := termbox.ColorBlack
		switch nPegs {
		case 0:
		case 1:
			background = termbox.ColorBlue
		case 2:
			background = termbox.ColorGreen
		default:
			background = termbox.ColorRed
		}
		screenPt := pt.Sub(min)
		termbox.SetCell(screenPt.X, screenPt.Y, rune(cell),
			foreground, background)
	}

	if pegIdx < len(gen.OpenPegs()) {
		fits := []chunk.OffsetChunk{}
		// Add the chunks that fit in the lattice
		for _, oc := range gen.FittingChunks(gen.OpenPegs()[pegIdx], chunks) {
			fits = append(fits, oc)
		}
		if len(fits) == 0 {
			gen.ClosePeg(gen.OpenPegs()[pegIdx])
		} else {
			oc := fits[chunkIdx%len(fits)]
			for y := oc.Bounds().Min.Y; y < oc.Bounds().Max.Y; y++ {
				for x := oc.Bounds().Min.X; x < oc.Bounds().Max.X; x++ {
					pt := image.Pt(x, y)
					screenPt := pt.Sub(min)
					if c, ok := oc.At(pt); ok {
						termbox.SetCell(screenPt.X, screenPt.Y, rune(c),
							termbox.ColorRed, termbox.ColorBlack)
					}
				}
			}
		}
	}

	termbox.Flush()
}

func spawn(gen *chunk.Gen, chunks []*chunk.Chunk, pegIdx int, chunkIdx int) {
	for {
		pegs := gen.OpenPegs()
		if len(pegs) == 0 {
			return
		}
		peg := pegs[pegIdx%len(pegs)]
		fits := []chunk.OffsetChunk{}
		// Add the chunks that fit in the lattice
		for _, oc := range gen.FittingChunks(peg, chunks) {
			fits = append(fits, oc)
		}
		if len(fits) == 0 {
			gen.ClosePeg(peg)
			continue
		}
		gen.AddChunk(fits[chunkIdx%len(fits)])
		return
	}
}

func main() {
	for _, ch := range strings.Split(chunkFile, "\n\n") {
		c, err := chunk.Parse(chunk.ParseSpec{"|.", '*'}, ch)
		if err != nil {
			panic(err)
		}
		chunks = append(chunks, c)
	}
	chunks = chunk.GenerateVariants(chunks)

	gen := chunk.New(chunks[0], '#')
	gen.SetGrid(image.Pt(4, 4))

	err := termbox.Init()
	if err != nil {
		panic(err)
	}
	defer termbox.Close()

	pegIdx := rand.Intn(len(gen.OpenPegs()))
	chunkIdx := 0
loop:
	for {
		draw(gen, pegIdx, chunkIdx)
		switch ev := termbox.PollEvent(); ev.Type {
		case termbox.EventKey:
			switch ev.Key {
			case termbox.KeyTab:
				pegIdx++
				chunkIdx = 0
				pegs := gen.OpenPegs()
				if pegIdx >= len(pegs) {
					pegIdx = 0
				}
			case termbox.KeySpace:
				chunkIdx++
			case termbox.KeyEnter:
				spawn(gen, chunks, pegIdx, chunkIdx)
				if len(gen.OpenPegs()) == 0 {
					pegIdx = 0
				} else {
					pegIdx = rand.Intn(len(gen.OpenPegs()))
				}
				chunkIdx = rand.Intn(256)
			case termbox.KeyEsc:
				break loop
			}
		}
	}
}
