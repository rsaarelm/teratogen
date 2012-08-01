/* cursor.go

   Copyright (C) 2012 Risto Saarelma

   This program is free software: you can redistribute it and/or modify
   it under the terms of the GNU General Public License as published by
   the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.

   This program is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU General Public License for more details.

   You should have received a copy of the GNU General Public License
   along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

package font

import (
	"image"
	"image/color"
	"teratogen/gfx"
)

type Edge uint8

const (
	None Edge = iota
	Emboss
	DropShadow
	Round
	Blocky
)

func (e Edge) Drawn(x, y int) bool {
	switch e {
	case Emboss:
		return (x < 0 || y < 0) && (x <= 0 && y <= 0)
	case DropShadow:
		return x > 0 && y > 0
	case Round:
		return (x == 0) != (y == 0)
	case Blocky:
		return x != 0 || y != 0
	}
	return false
}

type Cursor struct {
	Font    *Font
	Target  gfx.Surface32Bit
	Pos     image.Point
	Edge    Edge
	Col     color.Color
	EdgeCol color.Color
	Scale   int // Edge scaling
}

func (c *Cursor) Write(p []byte) (n int, err error) {
	scale := c.Scale
	if scale == 0 {
		scale = 1
	}
	// edge logic
	for y := -1; y <= 1; y++ {
		for x := -1; x <= 1; x++ {
			if c.Edge.Drawn(x, y) {
				c.Font.RenderTo32Bit(string(p), c.EdgeCol, c.Pos.X+(x*scale), c.Pos.Y+(y*scale), c.Target)
			}
		}
	}
	xAdv := c.Font.RenderTo32Bit(string(p), c.Col, c.Pos.X, c.Pos.Y, c.Target)
	c.Pos = image.Pt(c.Pos.X+int(xAdv), c.Pos.Y)
	return len(p), nil
}
