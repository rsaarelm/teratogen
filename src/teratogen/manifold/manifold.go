/* manifold.go

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

// Package manifold provides tools for working with a game world with portals.
// The name "manifold" refers to the topological concept for a structure which
// looks like regular space when viewed around a specific location, but not
// when seen as a whole.
package manifold

import (
	"fmt"
	"image"
)

// A single point in the manifold. Zone value 0 denotes inactive portals, so
// it should not be used in any Locations in use. You can't make portals that
// go to locations in zone 0.
type Location struct {
	X, Y int8
	Zone uint16
}

type Portal Location

// Add returns a location translated by a vector. Does not know about portals.
func (loc Location) Add(vec image.Point) Location {
	return Location{loc.X + int8(vec.X), loc.Y + int8(vec.Y), loc.Zone}
}

// Beyond returns the location beyond the given portal from the current location.
func (loc Location) Beyond(portal Portal) Location {
	if portal.Zone != 0 {
		return Location{loc.X + portal.X, loc.Y + portal.Y, portal.Zone}
	}
	return loc
}

func (loc Location) String() string {
	return fmt.Sprintf("(%d: %d, %d)", loc.Zone, loc.X, loc.Y)
}

func (loc Portal) IsNull() bool {
	return loc.X == 0 && loc.Y == 0 && loc.Zone == 0
}

func (loc Portal) String() string {
	return fmt.Sprintf("->(%d: %d, %d)", loc.Zone, loc.X, loc.Y)
}

func NullPortal() Portal {
	return Portal{}
}

func Loc(x, y int8, zone uint16) Location {
	return Location{x, y, zone}
}

func Port(dx, dy int8, targetZone uint16) Portal {
	return Portal{dx, dy, targetZone}
}

// Chart is a mapping from a two-dimensional Euclidean plane into some set of
// locations in a manifold. A field of view of a game character from a
// specific origin location would produce a chart for that origin. The name
// refers to atlases and charts of topological manifold.
type Chart interface {
	At(pt image.Point) Location
}

type Manifold struct {
	portals map[Location]Portal
}

func New() *Manifold {
	return &Manifold{make(map[Location]Portal)}
}

// Offset returns a portaled location the vector away from the initial one.
// Only the portal exactly the vector's span away from the initial location
// matters; you will probably mostly want to use this with unit length
// vectors.
func (m *Manifold) Offset(loc Location, vec image.Point) (newLoc Location) {
	return m.Traverse(loc.Add(vec))
}

func (m *Manifold) Traverse(loc Location) Location {
	return loc.Beyond(m.Portal(loc))
}

func (m *Manifold) Portal(loc Location) Portal {
	if portal, ok := m.portals[loc]; ok {
		return portal
	}
	return NullPortal()
}

func (m *Manifold) SetPortal(loc Location, portal Portal) {
	m.portals[loc] = portal
}

func (m *Manifold) ClearPortal(loc Location) {
	delete(m.portals, loc)
}
