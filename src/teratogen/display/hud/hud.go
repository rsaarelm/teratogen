// hud.go
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

// Package hud handles displaying onscreen status readouts and messages during
// gameplay.
package hud

import (
	"fmt"
	"image"
	"teratogen/cache"
	"teratogen/display/util"
	"teratogen/entity"
	"teratogen/font"
	"teratogen/gfx"
	"teratogen/sdl"
	"teratogen/world"
)

type Hud struct {
	cache *cache.Cache
	world *world.World
}

func New(c *cache.Cache, w *world.World) *Hud {
	return &Hud{cache: c, world: w}
}

func (h *Hud) Draw(bounds image.Rectangle) {
	sdl.Frame().SetClipRect(bounds)
	defer sdl.Frame().ClearClipRect()

	f, err := h.cache.GetFont(font.Spec{"assets/BMmini.ttf", 8.0, 32, 96})
	if err != nil {
		panic(err)
	}
	cur := &font.Cursor{f, sdl.Frame(), bounds.Min.Add(image.Pt(0, int(f.Height()))),
		font.None, gfx.Yellow, gfx.Black}

	fmt.Fprintf(cur, "Heavy boxes perform quick waltzes and jigs.")

	h.drawHealth(image.Rectangle{bounds.Min.Add(image.Pt(0, 8)), bounds.Max})
}

func (h *Hud) drawHealth(bounds image.Rectangle) {
	heart := h.cache.GetDrawable(util.SmallIcon(util.Items, 22))
	halfHeart := h.cache.GetDrawable(util.SmallIcon(util.Items, 23))
	noHeart := h.cache.GetDrawable(util.SmallIcon(util.Items, 24))
	shield := h.cache.GetDrawable(util.SmallIcon(util.Items, 25))
	halfShield := h.cache.GetDrawable(util.SmallIcon(util.Items, 26))

	pc, _ := h.world.Player.(entity.Stats)
	offset := bounds.Min
	for i := 0; i < pc.MaxHealth(); i += 2 {
		n := pc.Health() - i
		if n > 1 {
			heart.Draw(offset)
		} else if n > 0 {
			halfHeart.Draw(offset)
		} else {
			noHeart.Draw(offset)
		}
		offset = offset.Add(image.Pt(util.TileW, 0))
	}
	for i := 0; i < pc.Shield(); i += 2 {
		n := pc.Shield() - i
		if n > 1 {
			shield.Draw(offset)
		} else {
			halfShield.Draw(offset)
		}
		offset = offset.Add(image.Pt(util.TileW, 0))
	}
}
