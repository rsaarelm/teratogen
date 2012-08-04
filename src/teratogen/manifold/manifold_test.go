// manifold_test.go
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

package manifold

import (
	"image"
	"testing"
)

func TestSimpleManifold(t *testing.T) {
	spc := New()

	origin := Loc(0, 0, 1)

	if spc.Portal(origin) != NullPortal() {
		t.Fail()
	}

	spc.SetPortal(Loc(1, 0, 1), Port(10, 10, 2))
	b := Loc(0, 0, 0)

	if spc.Offset(origin, image.Pt(1, 0)) == b {
		t.Fail()
	}

	if spc.Offset(origin, image.Pt(1, 0)) != Loc(11, 10, 2) {
		t.Fail()
	}

	if spc.Offset(origin, image.Pt(0, 1)) != Loc(0, 1, 1) {
		t.Fail()
	}

	spc.ClearPortal(Loc(1, 0, 1))

	if spc.Offset(origin, image.Pt(1, 0)) != Loc(1, 0, 1) {
		t.Fail()
	}
}

func TestFootprint(t *testing.T) {
	spc := New()

	emptyShape := []image.Point{}
	goodShape := []image.Point{{1, 0}, {3, 0}, {0, 1}, {-1, 0}, {1, 1}, {2, 0}}
	badShape := []image.Point{{1, 0}, {10, 0}}

	if template, err := MakeTemplate(emptyShape); err == nil {
		footprint := spc.MakeFootprint(template, Loc(0, 0, 1))
		if len(footprint) != 1 {
			t.Fail()
		}
		if _, ok := footprint[image.Pt(0, 0)]; !ok {
			t.Fail()
		}
	} else {
		t.Fail()
	}

	if template, err := MakeTemplate(goodShape); err == nil {
		// Test the multi-cell footprint first in basic manifold.

		spc = New()

		footprint := spc.MakeFootprint(template, Loc(0, 0, 1))
		if footprint[image.Pt(2, 0)] != Loc(2, 0, 1) {
			t.Fail()
		}

		// Now add a portal the footprint should spread through.

		spc.SetPortal(Loc(1, 0, 1), Port(20, 20, 2))
		footprint = spc.MakeFootprint(template, Loc(0, 0, 1))

		if footprint[image.Pt(2, 0)] != Loc(22, 20, 2) {
			t.Fail()
		}
	} else {
		t.Fail()
	}

	if _, err := MakeTemplate(badShape); err == nil {
		t.Fail()
	}
}
