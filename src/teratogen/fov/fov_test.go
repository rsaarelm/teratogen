// fov_test.go
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

package fov

import (
	"image"
	"teratogen/space"
	"testing"
)

func TestFov(t *testing.T) {
	mf := space.New()

	mf.SetPortal(space.Loc(10, 11, 1), space.Port(20, 20, 20))
	seen := map[image.Point]space.Location{}

	// Impassable barrier on every zone at x == 11
	blockFn := func(loc space.Location) bool { return loc.X == 11 }
	markFn := func(pt image.Point, loc space.Location) {
		seen[pt] = loc
	}

	fov := New(blockFn, markFn, mf)

	fov.Run(space.Loc(10, 10, 1), 4)

	if _, ok := seen[image.Pt(-2, 0)]; !ok {
		t.Fail()
	}

	if _, ok := seen[image.Pt(2, 0)]; ok {
		// Should be stopped by blockFn barrier
		t.Fail()
	}

	if loc, ok := seen[image.Pt(0, 2)]; ok {
		// This one should be through the portal.
		if loc.Zone != 20 {
			t.Fail()
		}
	} else {
		t.Fail()
	}
}
