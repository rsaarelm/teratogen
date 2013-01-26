// entity.go
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

// Package entity provides interfaces for the game world entities, such as
// items and creatures.
package entity

import (
	"image"
	"teratogen/space"
)

// BlockMove is an entity that can block movement.
type BlockMove interface {
	BlocksMove() bool
}

// Fov is a field of view component, it means an entity can remember the
// surroundings it has seen in a manifold chart.
type Fov interface {
	FovChart() space.Chart
	MoveFovOrigin(vec image.Point)
	MarkFov(pt image.Point, loc space.Location)
	ClearFov()
}

// Stats are the interface for active entities that fight and get hurt.
type Stats interface {
	Health() int
	MaxHealth() int
	Shield() int
	Damage(amount int)
	AddHealth(amount int)
	AddShield(amount int)
}

// Entity type is just an alias for interface{} for more explicit notation.
type Entity interface{}
