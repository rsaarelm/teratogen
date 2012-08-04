// display_test.go
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
	"math/rand"
	"testing"
	"testing/quick"
)

func validPoint(ix, iy int16) bool {
	x, y := int(ix), int(iy)
	chart1 := image.Pt(x, y)
	screen := ChartToScreen(chart1)
	chart2 := ScreenToChart(screen)
	// All points within tile rect should fall into chart.
	tileOffset := image.Pt(rand.Intn(TileW), rand.Intn(TileH))
	chart3 := ScreenToChart(screen.Add(tileOffset))

	fmt.Println(chart1, screen)
	fmt.Println(chart2)
	fmt.Println(chart3, tileOffset)
	if chart1 != chart3 {
		fmt.Println(tileOffset)
		return false
	}

	return chart1 == chart2
}

func TestCoordTransform(t *testing.T) {
	if err := quick.Check(validPoint, nil); err != nil {
		t.Error(err)
	}
}
