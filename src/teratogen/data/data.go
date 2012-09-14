// data.go
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

// Package data contains structured data definitions for the game world and
// rules.
package data

import (
	"image"
	"teratogen/gfx"
	"teratogen/mob"
)

var PcPortrait = []gfx.ImageSpec{
	{"assets/chars.png", image.Rect(0, 104, 24, 128)},
	{"assets/chars.png", image.Rect(24, 104, 48, 128)},
	{"assets/chars.png", image.Rect(48, 104, 72, 128)},
}

var PcDescr = []string{
	"MECHANIST",
	"BIOROID",
	"WARPER",
}

func NumClasses() int {
	return len(PcPortrait)
}

var PcSpec = []mob.Spec{
	{gfx.ImageSpec{"assets/chars.png", image.Rect(0, 16, 8, 24)}},
	{gfx.ImageSpec{"assets/chars.png", image.Rect(8, 16, 16, 24)}},
	{gfx.ImageSpec{"assets/chars.png", image.Rect(16, 16, 24, 24)}},
}
