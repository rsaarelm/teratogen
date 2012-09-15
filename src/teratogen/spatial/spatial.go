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
	"teratogen/entity"
	"teratogen/manifold"
)

type Spatial struct {
	manifold  *manifold.Manifold
	placement map[interface{}]manifold.Footprint
	sites     map[manifold.Location]siteSet
}

func New(m *manifold.Manifold) (result *Spatial) {
	result = new(Spatial)
	result.manifold = m
	result.placement = make(map[interface{}]manifold.Footprint)
	result.sites = make(map[manifold.Location]siteSet)
	return
}

// AddFootprints adds an entity with a custom, multi-cell footprint to the
// spatial index.
func (s *Spatial) AddFootprint(
	e interface{}, footprint manifold.Footprint) {
	if _, ok := s.placement[e]; ok {
		s.Remove(e)
	}

	s.placement[e] = footprint

	for offset, siteLoc := range footprint {
		s.initSite(siteLoc)
		s.sites[siteLoc][OffsetEntity{e, offset}] = true
	}
}

// Add adds an entity with a default single-cell footprint at loc. Entities
// that implement the entity.Footprint interface will get an extended
// footprint.
func (s *Spatial) Add(e interface{}, loc manifold.Location) {
	s.AddFootprint(e, s.EntityFootprint(e, loc))
}

func (s *Spatial) EntityFootprint(e interface{}, loc manifold.Location) manifold.Footprint {
	foot := manifold.Footprint{image.Pt(0, 0): loc}
	if ft, ok := e.(entity.Footprint); ok {
		foot = s.manifold.MakeFootprint(ft.Footprint(), loc)
	}
	return foot
}

func (s *Spatial) Contains(e interface{}) bool {
	_, ok := s.placement[e]
	return ok
}

func (s *Spatial) Loc(e interface{}) manifold.Location {
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

func (s *Spatial) At(loc manifold.Location) (result []OffsetEntity) {
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
	Entity interface{}
	Offset image.Point
}
