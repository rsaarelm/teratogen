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
	"image"
	"math"
	"teratogen/manifold"
	"teratogen/world"
)

type Display struct {
	world *world.World
	chart manifold.Chart
}

const TileW = 8
const TileH = 8

func floor(i float64) int {
	return int(math.Floor(i))
}

func ChartToScreen(chartPt image.Point) (screenPt image.Point) {
	return image.Pt(chartPt.X*TileW-chartPt.Y*TileW,
		chartPt.X*TileH/2+chartPt.Y*TileH/2)
}

func ScreenToChart(screenPt image.Point) (chartPt image.Point) {
	column := int(math.Floor(float64(screenPt.X) / TileW))
	row := int(math.Floor(float64(screenPt.Y-column*(TileH/2)) / TileH))
	return image.Pt(column+row, row)
}

func (d Display) Draw(offset image.Point) {

}
