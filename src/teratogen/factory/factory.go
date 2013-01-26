// factory.go
//
// Copyright (C) 2013 Risto Saarelma
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

// Package factory contains data and utilities for generating game entities.
package factory

import (
	"math/rand"
	"teratogen/display/util"
	"teratogen/entity"
	"teratogen/gfx"
	"teratogen/mob"
	"teratogen/world"
)

type spawnFunc func(*world.World) entity.Entity

type spawn struct {
	commonness int
	minDepth   int
	init       spawnFunc
}

func genPC(w *world.World) entity.Entity {
	return mob.NewPC(w, mob.Spec{util.SmallIcon(util.Chars, 16), 6})
}

func icon(idx int) gfx.ImageSpec { return util.SmallIcon(util.Chars, idx) }

func pc(icon gfx.ImageSpec, health int) spawnFunc {
	return spawnFunc(func(w *world.World) entity.Entity {
		return mob.NewPC(w, mob.Spec{icon, health})
	})
}

func monster(icon gfx.ImageSpec, health int) spawnFunc {
	return spawnFunc(func(w *world.World) entity.Entity {
		return mob.New(w, mob.Spec{icon, health})
	})
}

var spawns = map[string]spawn{
	"player":     {0, 0, pc(icon(16), 20)},
	"zombie":     {30, 0, monster(icon(1), 2)},
	"dog-thing":  {40, 0, monster(icon(2), 1)},
	"spitter":    {15, 2, monster(icon(3), 2)},
	"cyclops":    {15, 2, monster(icon(6), 2)},
	"death ooze": {15, 3, monster(icon(7), 4)},
	"bear":       {3, 0, monster(icon(23), 4)},
}

const (
	Player = "player"
)

func Spawn(id string, w *world.World) entity.Entity {
	if spawn, ok := spawns[id]; ok {
		return spawn.init(w)
	}
	panic("Unknown spawn id")
}

func RandomMonster(depth int, w *world.World) entity.Entity {
	dist := map[string]int{}
	total := 0
	for name, s := range spawns {
		if s.minDepth <= depth && s.commonness > 0 {
			dist[name] = s.commonness
			total += s.commonness
		}
	}

	if len(dist) == 0 {
		panic("Empty distribution for random spawn")
	}

	x := rand.Intn(total)
	for name, weight := range dist {
		x -= weight
		if x <= 0 {
			return spawns[name].init(w)
		}
	}
	panic("Random spawn failed")
}
