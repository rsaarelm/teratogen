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
	"image"
	"teratogen/app"
	"teratogen/display/util"
	"teratogen/entity"
	"teratogen/gfx"
	"teratogen/sdl"
	"teratogen/typography"
	"teratogen/world"
	"time"
)

const timeToReadLetter = .05e9

type Hud struct {
	world *world.World

	msgs       []string
	msgExpires int64
}

func New(w *world.World) *Hud {
	return &Hud{world: w}
}

func (h *Hud) Draw(bounds image.Rectangle) {
	h.update()

	sdl.Frame().SetClipRect(bounds)
	defer sdl.Frame().ClearClipRect()

	style := util.TextStyle().ForeColor(gfx.Khaki).Edge(typography.Round)

	for lineY, str := range h.msgs {
		pos := bounds.Min.Add(image.Pt(0, (lineY+1)*int(style.LineHeight())))
		style.Render(str, pos)
	}

	h.drawHealth(image.Rectangle{image.Pt(bounds.Min.X, bounds.Max.Y-8), bounds.Max})
}

func (h *Hud) drawHealth(bounds image.Rectangle) {
	heart := app.Cache().GetDrawable(util.SmallIcon(util.Items, 22))
	halfHeart := app.Cache().GetDrawable(util.SmallIcon(util.Items, 23))
	noHeart := app.Cache().GetDrawable(util.SmallIcon(util.Items, 24))
	shield := app.Cache().GetDrawable(util.SmallIcon(util.Items, 25))
	halfShield := app.Cache().GetDrawable(util.SmallIcon(util.Items, 26))

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

func (h *Hud) Msg(str string) {
	h.msgs = append(h.msgs, str)
	if len(h.msgs) == 1 {
		h.setExpires()
	}
}

func (h *Hud) update() {
	t := time.Now().UnixNano()
	if len(h.msgs) > 0 {
		if t >= h.msgExpires {
			// Assume the oldest message is read and remove it.
			h.msgs = h.msgs[1:len(h.msgs)]
			h.setExpires()
		}
	}
}

func (h *Hud) setExpires() {
	if len(h.msgs) > 0 {
		delay := timeToReadLetter * int64(len(h.msgs[0]))
		if delay < 1e9 {
			delay = 1e9
		}
		h.msgExpires = time.Now().UnixNano() + delay
	}
}
