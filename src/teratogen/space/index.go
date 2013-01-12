// index.go
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

package space

import (
	"image"
)

// Index is a spatial index for indexing single and multi cell entities in
// space.
type Index struct {
	placement map[interface{}]Footprint
	sites     map[Location]siteSet
}

func NewIndex() (result *Index) {
	result = new(Index)
	result.Init()
	return
}

func (s *Index) Init() {
	s.placement = make(map[interface{}]Footprint)
	s.sites = make(map[Location]siteSet)
}

func (s *Index) Clear() {
	s.Init()
}

// Place places an entity with a custom, multi-cell footprint to the spatial
// index. If the entity has been previously placed in the index, it is removed
// before being placed into the new location.
func (s *Index) Place(
	e interface{}, footprint Footprint) {
	if _, ok := s.placement[e]; ok {
		s.Remove(e)
	}

	s.placement[e] = footprint

	for offset, siteLoc := range footprint {
		s.initSite(siteLoc)
		s.sites[siteLoc][OffsetEntity{e, offset}] = true
	}
}

func (s *Index) Contains(e interface{}) bool {
	_, ok := s.placement[e]
	return ok
}

func (s *Index) Loc(e interface{}) Location {
	return s.placement[e][image.Pt(0, 0)]
}

func (s *Index) ForEach(fn func(interface{})) {
	for e, _ := range s.placement {
		fn(e)
	}
}

func (s *Index) Remove(e interface{}) {
	footprint, ok := s.placement[e]
	if !ok {
		panic("Removing an unknown entity from spatial index")
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

func (s *Index) At(loc Location) (result []OffsetEntity) {
	site, ok := s.sites[loc]
	if !ok {
		return
	}
	for elt, _ := range site {
		result = append(result, elt)
	}
	return
}

func (s *Index) initSite(loc Location) {
	if _, ok := s.sites[loc]; !ok {
		s.sites[loc] = make(siteSet)
	}
}

type siteSet map[OffsetEntity]bool

type OffsetEntity struct {
	Entity interface{}
	Offset image.Point
}
