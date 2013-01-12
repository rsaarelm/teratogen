// spatial.go
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

// Package spatial provides spatial indexing for entities in a manifold.
package spatial

import (
	"image"
	"teratogen/space"
)

type Spatial struct {
	placement map[interface{}]space.Footprint
	sites     map[space.Location]siteSet
}

func New() (result *Spatial) {
	result = new(Spatial)
	result.Init()
	return
}

func (s *Spatial) Init() {
	s.placement = make(map[interface{}]space.Footprint)
	s.sites = make(map[space.Location]siteSet)
}

func (s *Spatial) Clear() {
	s.Init()
}

// Place places an entity with a custom, multi-cell footprint to the spatial
// index. If the entity has been previously placed in the index, it is removed
// before being placed into the new location.
func (s *Spatial) Place(
	e interface{}, footprint space.Footprint) {
	if _, ok := s.placement[e]; ok {
		s.Remove(e)
	}

	s.placement[e] = footprint

	for offset, siteLoc := range footprint {
		s.initSite(siteLoc)
		s.sites[siteLoc][OffsetEntity{e, offset}] = true
	}
}

func (s *Spatial) Contains(e interface{}) bool {
	_, ok := s.placement[e]
	return ok
}

func (s *Spatial) Loc(e interface{}) space.Location {
	return s.placement[e][image.Pt(0, 0)]
}

func (s *Spatial) ForEach(fn func(interface{})) {
	for e, _ := range s.placement {
		fn(e)
	}
}

func (s *Spatial) Remove(e interface{}) {
	footprint, ok := s.placement[e]
	if !ok {
		panic("Removing an unknown entity from Spatial")
	}

top:
	for _, loc := range footprint {
		site := s.sites[loc]
		for sited, _ := range site {
			if sited.Entity == e {
				delete(site, sited)
				if len(site) == 0 {
					delete(s.sites, loc)
				}
				continue top
			}
		}
		panic("Entity not found on site belonging to footprint.")
	}
	delete(s.placement, e)
}

func (s *Spatial) At(loc space.Location) (result []OffsetEntity) {
	site, ok := s.sites[loc]
	if !ok {
		return
	}
	for elt, _ := range site {
		result = append(result, elt)
	}
	return
}

func (s *Spatial) initSite(loc space.Location) {
	if _, ok := s.sites[loc]; !ok {
		s.sites[loc] = make(siteSet)
	}
}

type siteSet map[OffsetEntity]bool

type OffsetEntity struct {
	Entity interface{}
	Offset image.Point
}
