// fov.go
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

package mob

import (
	"image"
	"teratogen/manifold"
)

// A field of view for mobs.
type Fov struct {
	relativePos image.Point
	chart       map[image.Point]manifold.Location
}

func NewFov() (result *Fov) {
	result = new(Fov)
	result.Init()
	return
}

func (f *Fov) Init() {
	f.chart = make(map[image.Point]manifold.Location)
}

// Use a separate type for the chart since chart's main method name "At" is
// too generic to embed straight into an entity.

type fovChart Fov

func (f *fovChart) At(pt image.Point) manifold.Location {
	if loc, ok := f.chart[pt.Add(f.relativePos)]; ok {
		return loc
	}
	return manifold.Location{}
}

func (f *Fov) FovChart() manifold.Chart {
	return (*fovChart)(f)
}

func (f *Fov) MarkFov(pt image.Point, loc manifold.Location) {
	f.chart[pt.Add(f.relativePos)] = loc
}

func (f *Fov) MoveFovOrigin(vec image.Point) {
	f.relativePos = f.relativePos.Add(vec)
}

func (f *Fov) ClearFov() {
	f.Init()
}
