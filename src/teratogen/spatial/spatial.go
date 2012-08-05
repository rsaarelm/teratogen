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

package spatial

import (
	"image"
	"teratogen/manifold"
)

type Entity interface {
	Footprint(loc manifold.Location) manifold.Footprint
}

type Spatial struct {
	placement map[Entity]manifold.Footprint
	sites     map[manifold.Location]siteSet
}

func New() (result *Spatial) {
	result = new(Spatial)
	result.placement = make(map[Entity]manifold.Footprint)
	result.sites = make(map[manifold.Location]siteSet)
	return
}

func (s *Spatial) Add(e Entity, loc manifold.Location) {
	if _, ok := s.placement[e]; ok {
		panic("Adding same entity multiple times to Spatial")
	}

	footprint := e.Footprint(loc)
	s.placement[e] = footprint

	for offset, siteLoc := range footprint {
		s.initSite(siteLoc)
		s.sites[siteLoc][OffsetEntity{e, offset}] = true
	}
}

func (s *Spatial) Remove(e Entity) {
	footprint, ok := s.placement[e]
	if !ok {
		panic("Removing an unknown entity from Spatial")
	}

	for _, loc := range footprint {
		site := s.sites[loc]
		for sited, _ := range site {
			if sited.Entity == e {
				delete(site, sited)
				if len(site) == 0 {
					delete(s.sites, loc)
				}
				break
			}
		}
		panic("Entity not found on site belonging to footprint.")
	}
}

func (s *Spatial) Get(loc manifold.Location) (result []OffsetEntity) {
	site, ok := s.sites[loc]
	if !ok {
		return
	}
	for elt, _ := range site {
		result = append(result, elt)
	}
	return
}

func (s *Spatial) initSite(loc manifold.Location) {
	if _, ok := s.sites[loc]; !ok {
		s.sites[loc] = make(siteSet)
	}
}

type siteSet map[OffsetEntity]bool

type OffsetEntity struct {
	Entity Entity
	Offset image.Point
}
