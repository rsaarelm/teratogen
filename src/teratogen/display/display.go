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
	"teratogen/cache"
	"teratogen/display/view"
	"teratogen/font"
	"teratogen/gfx"
	"teratogen/sdl"
	"teratogen/world"
)

type Display struct {
	cache       *cache.Cache
	view        *view.View
	chartOrigin image.Point
}

func New(c *cache.Cache, w *world.World) (result *Display) {
	result = new(Display)
	result.cache = c
	result.view = view.New(c, w)

	return
}

func (d *Display) DrawWorld(bounds image.Rectangle) {
	sdl.Frame().SetClipRect(bounds)
	defer sdl.Frame().ClearClipRect()
	sdl.Frame().Clear(gfx.Black)

	sprites := gfx.SpriteBatch{}

	sprites = d.view.CollectSprites(sprites, bounds)

	sprites.Sort()
	sprites.Draw()
}

func (d *Display) DrawMsg(bounds image.Rectangle) {
	sdl.Frame().SetClipRect(bounds)
	defer sdl.Frame().ClearClipRect()
	sdl.Frame().Clear(gfx.Black)

	f, err := d.cache.GetFont(font.Spec{"assets/BMmini.ttf", 8.0, 32, 96})
	if err != nil {
		panic(err)
	}
	cur := &font.Cursor{f, sdl.Frame(), bounds.Min.Add(image.Pt(0, int(f.Height()))),
		font.None, gfx.Green, gfx.Black}

	fmt.Fprintf(cur, "Heavy boxes perform quick waltzes and jigs.")
}
