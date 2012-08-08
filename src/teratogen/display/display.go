// display.go
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

package display

import (
	"fmt"
	"image"
	"math"
	"teratogen/cache"
	"teratogen/entity"
	"teratogen/font"
	"teratogen/gfx"
	"teratogen/manifold"
	"teratogen/sdl"
	"teratogen/tile"
	"teratogen/world"
)

type Display struct {
	cache       *cache.Cache
	world       *world.World
	chart       *world.FovChart
	chartOrigin image.Point
}

func New(c *cache.Cache, w *world.World) (result *Display) {
	result = new(Display)
	result.cache = c
	result.world = w

	result.chart = world.NewFov(w)
	// XXX: Magic number fov radius
	result.chart.DoFov(manifold.Loc(0, 0, 1), 12)

	return
}

const TileW = 8
const TileH = 8

func ChartToScreen(chartPt image.Point) (scrPt image.Point) {
	return image.Pt(chartPt.X*TileW-chartPt.Y*TileW,
		chartPt.X*TileH/2+chartPt.Y*TileH/2)
}

func ScreenToChart(scrPt image.Point) (chartPt image.Point) {
	column := int(math.Floor(float64(scrPt.X) / TileW))
	row := int(math.Floor(float64(scrPt.Y-column*(TileH/2)) / TileH))
	return image.Pt(column+row, row)
}

func (d *Display) Move(vec image.Point) {
	d.chart.Move(vec)

	// XXX: HACK
	d.chart.DoFov(manifold.Loc(int8(d.chart.RelativePos.X), int8(d.chart.RelativePos.Y), 1), 12)
}

func (d *Display) drawCell(chartPos image.Point, scrPos image.Point) {
	loc := d.chart.At(chartPos)
	if d.world.Contains(loc) {
		idx := terrainTileOffset(d.world, d.chart, chartPos)
		sprite := d.cache.GetDrawable(d.world.Terrain(loc).Icon[idx])
		sprite.Draw(scrPos)
	}

	// XXX: Totally hacked player sprite placement, replace with proper entity
	// sprites.
	if chartPos == image.Pt(0, 0) {
		pcSprite := d.cache.GetDrawable(gfx.ImageSpec{"assets/chars.png", image.Rect(0, 8, 8, 16)})
		pcSprite.Draw(scrPos)
	}
}

func (d *Display) collectSprites(
	sprites gfx.SpriteBatch,
	chartPos image.Point) gfx.SpriteBatch {
	loc := d.chart.At(chartPos)

	// Collect terrain tile sprite.
	if d.world.Contains(loc) {
		idx := terrainTileOffset(d.world, d.chart, chartPos)
		sprites = append(sprites, gfx.Sprite{
			Layer:    entity.TerrainLayer,
			Drawable: d.cache.GetDrawable(d.world.Terrain(loc).Icon[idx]),
			Offset:   d.chartToScreenPos(chartPos)})
	}

	// Collect dynamic object sprites.
	for _, oe := range d.world.Spatial.Get(loc) {
		spritable := oe.Entity.(gfx.Spritable)
		if spritable == nil {
			continue
		}
		objChartPos := chartPos.Sub(oe.Offset)
		sprites = append(
			sprites,
			spritable.Sprite(d.cache, d.chartToScreenPos(objChartPos)))
	}

	return sprites
}

func corners(rect image.Rectangle) []image.Point {
	return []image.Point{rect.Min, rect.Max,
		{rect.Min.X, rect.Max.Y}, {rect.Max.X, rect.Min.Y}}
}

func centerOrigin(screenArea image.Rectangle) (screenPos image.Point) {
	return screenArea.Min.Add(screenArea.Size().Div(2)).Sub(image.Pt(TileW/2, TileH/2))
}

// chartArea returns the smallest rectangle containing all chart positions
// that can get drawn in the given screen rectangle, if chart position (0, 0)
// is at the center of the screen rectangle.
func ChartArea(screenArea image.Rectangle) image.Rectangle {
	scrOrigin := centerOrigin(screenArea)
	minX, minY := math.MaxInt32, math.MaxInt32
	maxX, maxY := math.MinInt32, math.MinInt32
	for _, pt := range corners(screenArea.Sub(scrOrigin)) {
		chartPos := ScreenToChart(pt)
		minX = int(math.Min(float64(chartPos.X), float64(minX)))
		minY = int(math.Min(float64(chartPos.Y), float64(minY)))
		maxX = int(math.Max(float64(chartPos.X), float64(maxX)))
		maxY = int(math.Max(float64(chartPos.Y), float64(maxY)))
	}

	return image.Rect(minX, minY, maxX+1, maxY+1)
}

func (d *Display) chartToScreenPos(chartPos image.Point) image.Point {
	return ChartToScreen(chartPos).Add(d.chartOrigin)
}

func (d *Display) DrawWorld(bounds image.Rectangle) {
	sdl.Frame().SetClipRect(bounds)
	defer sdl.Frame().ClearClipRect()
	sdl.Frame().Clear(gfx.Black)

	chartBounds := ChartArea(bounds)
	d.chartOrigin = centerOrigin(bounds)

	sprites := gfx.SpriteBatch{}

	for y := chartBounds.Min.Y; y < chartBounds.Max.Y; y++ {
		for x := chartBounds.Min.X; x < chartBounds.Max.X; x++ {
			sprites = d.collectSprites(sprites, image.Pt(x, y))
		}
	}

	// XXX: Hack to show up a fake player avatar.
	sprites = append(sprites, gfx.Sprite{
		Layer:    entity.MobLayer,
		Drawable: d.cache.GetDrawable(gfx.ImageSpec{"assets/chars.png", image.Rect(0, 8, 8, 16)}),
		Offset:   d.chartToScreenPos(image.Pt(0, 0))})

	sprites.Sort()
	sprites.Draw()
}

func (d *Display) DrawMsg(bounds image.Rectangle) {
	sdl.Frame().SetClipRect(bounds)
	defer sdl.Frame().ClearClipRect()
	sdl.Frame().Clear(gfx.Black)

	f, err := d.cache.GetFont(font.Spec{"assets/04round_bold.ttf", 8.0, 32, 96})
	if err != nil {
		panic(err)
	}
	cur := &font.Cursor{f, sdl.Frame(), bounds.Min.Add(image.Pt(0, int(f.Height()))),
		font.None, gfx.Green, gfx.Black}

	fmt.Fprintf(cur, "Heavy boxes perform quick waltzes and jigs.")
}

// terrainTileOffest checks the neighbourhood of a charted tile to see if it
// needs special formatting. Mostly used to prettify wall tiles.
func terrainTileOffset(w *world.World, chart manifold.Chart, pos image.Point) int {
	t := w.Terrain(chart.At(pos))
	if t.Kind == world.WallKind {
		edgeMask := 0
		for i, vec := range tile.HexDirs {
			if w.Terrain(chart.At(pos.Add(vec))).ShapesWalls() {
				edgeMask |= 1 << uint8(i)
			}
		}
		return tile.HexWallType(edgeMask)
	}
	return 0
}
