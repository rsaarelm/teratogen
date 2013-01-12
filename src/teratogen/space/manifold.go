// manifold.go
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

// Package space provides tools for working with a game world with portals.
package space

import (
	"fmt"
	"image"
)

// Location is a single point in space. Zone value 0 denotes inactive portals.
// Since this means you can't make portals that go to locations in zone 0, you
// should never have actual locations in zone 0. By convention, the default
// value Location{0, 0, 0} means "no place" and can be used to denote an
// invalid location.
type Location struct {
	X, Y int8
	Zone uint16
}

// Portal has the same structure as a location, but it's X and Y fields
// indicate relative displacement caused by moving through the portal and the
// Zone (absolute value, unlike X and Y) indicates the zone where the portal
// will lead. Portals whose Zone is 0 are treated as inactive.
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

func (loc Portal) String() string {
	return fmt.Sprintf("->(%d: %d, %d)", loc.Zone, loc.X, loc.Y)
}

// NullPortal returns the default value for a portal that doesn't go anywhere.
// It is used to represent a location not having a portal.
func NullPortal() Portal {
	return Portal{}
}

// Loc is a convenience function for creating location values.
func Loc(x, y int8, zone uint16) Location {
	return Location{x, y, zone}
}

// Port is a convenience function for creating portal values.
func Port(dx, dy int8, targetZone uint16) Portal {
	return Portal{dx, dy, targetZone}
}

// Chart is a mapping from a two-dimensional Euclidean plane into some set of
// locations in a manifold. A field of view of a game character from a
// specific origin location would produce a chart for that origin. The name
// refers to atlases and charts of topological manifold. All charts map the
// entire Euclidean plane. The default {0, 0, 0} location value is used for
// "undefined" points, such as points outside a field of view.
type Chart interface {
	At(pt image.Point) Location
}

// MapChart is a simple explicit map structure that implements the Chart interface.
type MapChart map[image.Point]Location

func (s MapChart) At(pt image.Point) Location {
	if loc, ok := map[image.Point]Location(s)[pt]; ok {
		return loc
	}
	return Location{}
}

// Manifold is the collection of portals that defines the structure of a game
// space. The name "manifold" refers to the topological concept for a
// structure which looks like regular space when viewed around a specific
// location, but not when seen as a whole.
type Manifold struct {
	portals map[Location]Portal
}

func NewManifold() *Manifold {
	return &Manifold{make(map[Location]Portal)}
}

// Offset returns a portaled location the vector away from the initial one.
// Only the portal exactly the vector's span away from the initial location
// matters; you will probably mostly want to use this with unit length
// vectors.
func (m *Manifold) Offset(loc Location, vec image.Point) (newLoc Location) {
	return m.Traverse(loc.Add(vec))
}

// Traverse returns the location beyond a portal at the argument location, if
// there is a portal, otherwise it returns the argument location.
func (m *Manifold) Traverse(loc Location) Location {
	return loc.Beyond(m.Portal(loc))
}

// Portal returns the portal at the argument location. A null portal will be
// returned for locations that do not have an explicit portal.
func (m *Manifold) Portal(loc Location) Portal {
	if portal, ok := m.portals[loc]; ok {
		return portal
	}
	return NullPortal()
}

// SetPortal sets the portal at the given location. If the portal value equals
// NullPortal, the explicit portal will be cleared from the manifold data
// structure.
func (m *Manifold) SetPortal(loc Location, portal Portal) {
	if portal == NullPortal() {
		delete(m.portals, loc)
	} else {
		m.portals[loc] = portal
	}
}
