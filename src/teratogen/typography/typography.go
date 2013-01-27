// typography.go
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

// Package typography handles printing text on a graphical display.
package typography

import (
	"image"
	"image/color"
	"teratogen/app"
	"teratogen/font"
	"teratogen/gfx"
	"teratogen/sdl"
)

type Style struct {
	font      *font.Font
	edge      Edge
	foreColor color.Color
	backColor color.Color
}

func NewStyle(fontSpec font.Spec) *Style {
	return &Style{
		font:      app.Cache().GetFont(fontSpec),
		edge:      None,
		foreColor: gfx.White,
		backColor: gfx.Black}
}

func (s *Style) Edge(edge Edge) *Style {
	result := *s
	result.edge = edge
	return &result
}

func (s *Style) ForeColor(col color.Color) *Style {
	result := *s
	result.foreColor = col
	return &result
}

func (s *Style) BackColor(col color.Color) *Style {
	result := *s
	result.backColor = col
	return &result
}

func (s *Style) Colors(foreColor, backColor color.Color) *Style {
	result := *s
	result.foreColor = foreColor
	result.backColor = backColor
	return &result
}

func (s *Style) RenderOn(target gfx.Surface32Bit, line string, pos image.Point) {
	// edge logic
	for y := -1; y <= 1; y++ {
		for x := -1; x <= 1; x++ {
			if s.edge.drawn(x, y) {
				s.font.RenderTo32Bit(line, s.backColor, pos.X+x, pos.Y+y, target)
			}
		}
	}
	// main string
	s.font.RenderTo32Bit(line, s.foreColor, pos.X, pos.Y, target)
}

func (s *Style) LineHeight() float64 {
	return s.font.Height()
}

func (s *Style) Render(line string, pos image.Point) {
	s.RenderOn(sdl.Frame(), line, pos)
}

func (s *Style) Bounds(line string) image.Rectangle {
	w := int(s.font.StringWidth(line))
	h := int(s.font.Height())
	return image.Rect(0, -h, w, 0)
}

type Edge uint8

const (
	None Edge = iota
	Emboss
	DropShadow
	Round
	Blocky
)

func (e Edge) drawn(x, y int) bool {
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
