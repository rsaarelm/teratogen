// view.go
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

// Package view handles drawing the persistent game world on screen.
package view

import (
	"image"
	"teratogen/cache"
	"teratogen/display/util"
	"teratogen/entity"
	"teratogen/gfx"
	"teratogen/space"
	"teratogen/world"
)

type View struct {
	cache *cache.Cache
	world *world.World
}

func New(c *cache.Cache, w *world.World) (result *View) {
	result = new(View)
	result.cache = c
	result.world = w
	return
}

func (v *View) chart() space.Chart {
	return v.world.Player.FovChart()
}

func (v *View) collectSpritesAt(
	sprites gfx.SpriteBatch,
	chartPos image.Point,
	screenOffset image.Point) gfx.SpriteBatch {
	loc := v.chart().At(chartPos)

	// Collect terrain tile sprite.
	if v.world.Contains(loc) {
		idx := util.TerrainTileOffset(v.world, v.chart(), chartPos)
		sprites = append(sprites, gfx.Sprite{
			Layer:    entity.TerrainLayer,
			Offset:   util.ChartToScreen(chartPos).Add(screenOffset),
			Drawable: v.cache.GetDrawable(v.world.Terrain(loc).Icon[idx])})
	}

	// Collect dynamic object sprites.
	for _, oe := range v.world.Spatial.At(loc) {
		spritable := oe.Entity.(gfx.Spritable)
		if spritable == nil {
			continue
		}
		objChartPos := chartPos.Sub(oe.Offset)
		sprites = append(
			sprites,
			spritable.Sprite(v.cache,
				util.ChartToScreen(objChartPos).Add(screenOffset)))
	}

	return sprites
}

// CollectSprites collects all the sprites in the visible world chart into the
// sprite collection.
func (v *View) CollectSprites(
	sprites gfx.SpriteBatch,
	bounds image.Rectangle) gfx.SpriteBatch {
	// TODO: Fog of war display on locations that are explored but not currently
	// visible.

	chartBounds := util.ChartArea(bounds)
	screenOffset := util.CenterOrigin(bounds)

	for y := chartBounds.Min.Y; y < chartBounds.Max.Y; y++ {
		for x := chartBounds.Min.X; x < chartBounds.Max.X; x++ {
			sprites = v.collectSpritesAt(sprites, image.Pt(x, y), screenOffset)
		}
	}

	return sprites
}
