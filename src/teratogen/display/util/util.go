// util.go
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

// Package util contains utility functions for the on-screen display.
package util

import (
	"image"
	"math"
	"teratogen/app"
	"teratogen/font"
	"teratogen/gfx"
)

const (
	TerrainLayer = 0
	DecalLayer   = 10
	ItemLayer    = 20
	MobLayer     = 30
	AnimLayer    = 100
)

const (
	TileW = 8
	TileH = 8
)

const (
	Tiles = "assets/tiles.png"
	Chars = "assets/chars.png"
	Items = "assets/items.png"
)

var HalfTile = image.Pt(TileW/2, TileH/2)

// ChartToScreen maps a point in the game tile coordinates into screen pixel
// coordinates that indicate where the tile should be drawn.
func ChartToScreen(chartPt image.Point) (scrPt image.Point) {
	return image.Pt(chartPt.X*TileW-chartPt.Y*TileW,
		chartPt.X*TileH/2+chartPt.Y*TileH/2)
}

// ScreenToChart maps a point in screen coordinates into the game tile chart
// coordinates for the tile on which the screen point falls on.
func ScreenToChart(scrPt image.Point) (chartPt image.Point) {
	column := int(math.Floor(float64(scrPt.X) / TileW))
	row := int(math.Floor(float64(scrPt.Y-column*(TileH/2)) / TileH))
	return image.Pt(column+row, row)
}

// Corners returns an array of the four corner points of a rectangle.
func Corners(rect image.Rectangle) []image.Point {
	return []image.Point{rect.Min, rect.Max,
		{rect.Min.X, rect.Max.Y}, {rect.Max.X, rect.Min.Y}}
}

// CenterOrigin returns the screen coordinates where a game tile that shows up
// at the center of the rectangle should be drawn.
func CenterOrigin(screenArea image.Rectangle) (screenPos image.Point) {
	return screenArea.Min.Add(screenArea.Size().Div(2)).Sub(image.Pt(TileW/2, TileH/2))
}

// ChartArea returns the smallest rectangle containing all chart positions
// that can get drawn in the given screen rectangle, if chart position (0, 0)
// is at the center of the screen rectangle.
func ChartArea(screenArea image.Rectangle) image.Rectangle {
	scrOrigin := CenterOrigin(screenArea)
	minX, minY := math.MaxInt32, math.MaxInt32
	maxX, maxY := math.MinInt32, math.MinInt32
	for _, pt := range Corners(screenArea.Sub(scrOrigin)) {
		chartPos := ScreenToChart(pt)
		minX = int(math.Min(float64(chartPos.X), float64(minX)))
		minY = int(math.Min(float64(chartPos.Y), float64(minY)))
		maxX = int(math.Max(float64(chartPos.X), float64(maxX)))
		maxY = int(math.Max(float64(chartPos.Y), float64(maxY)))
	}

	return image.Rect(minX, minY, maxX+1, maxY+1)
}

func smallIconRect(idx int) image.Rectangle {
	x, y := (idx%16)*TileW, (idx/16)*TileH
	return image.Rect(x, y, x+TileW, y+TileH)
}

func largeIconRect(idx int) image.Rectangle {
	x, y := (idx%5)*TileW*3, (idx/5)*TileH*3
	return image.Rect(x, y, x+TileW*3, y+TileH*3)
}

// SmallIcon returns a single-cell icon from the given icon sheet counting
// indexes from left to right and from top to bottom.
func SmallIcon(sheet string, idx int) gfx.ImageSpec {
	return gfx.SubImage(sheet, smallIconRect(idx))
}

// SmallIcons works like SmallIcon but produces multiple images.
func SmallIcons(sheet string, indices ...int) []gfx.ImageSpec {
	result := []gfx.ImageSpec{}
	for _, idx := range indices {
		result = append(result, SmallIcon(sheet, idx))
	}
	return result
}

// LargeIcon returns a sevel-cell icon (basically 3x3 small icons) from the
// given icon sheet counting indexes from left to right and top to bottom. The
// icon will be offseted so that it's draw position corresponds to its central
// cell.
func LargeIcon(sheet string, idx int) gfx.ImageSpec {
	return gfx.OffsetSubImage(sheet, largeIconRect(idx), image.Pt(-TileW, -TileH))
}

// LargeIcons works like LargeIcon but produces multiple images.
func LargeIcons(sheet string, indices ...int) []gfx.ImageSpec {
	result := []gfx.ImageSpec{}
	for _, idx := range indices {
		result = append(result, LargeIcon(sheet, idx))
	}
	return result
}

var defaultFont *font.Font = nil

func Font() *font.Font {
	if defaultFont == nil {
		var err error
		defaultFont, err = app.Cache().GetFont(font.Spec{"assets/BMmini.ttf", 8.0, 32, 96})
		if err != nil {
			panic(err)
		}
	}

	return defaultFont
}
