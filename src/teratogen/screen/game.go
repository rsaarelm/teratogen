// game.go
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

package screen

import (
	"image"
	"teratogen/action"
	"teratogen/app"
	"teratogen/data"
	"teratogen/display/anim"
	"teratogen/display/fx"
	"teratogen/display/hud"
	"teratogen/display/view"
	"teratogen/gfx"
	"teratogen/mapgen"
	"teratogen/mob"
	"teratogen/query"
	"teratogen/sdl"
	"teratogen/tile"
	"teratogen/world"
)

func Game() app.State {
	gs := new(game)
	return gs
}

type game struct {
	world  *world.World
	query  *query.Query
	hud    *hud.Hud
	view   *view.View
	anim   *anim.Anim
	fx     *fx.Fx
	action *action.Action
	mapgen *mapgen.Mapgen
}

func (gs *game) Enter() {
	gs.world = world.New()
	gs.query = query.New(gs.world)
	gs.hud = hud.New(app.Cache(), gs.world)
	gs.anim = anim.New()
	gs.view = view.New(app.Cache(), gs.world, gs.anim)
	gs.mapgen = mapgen.New(gs.world)
	gs.fx = fx.New(app.Cache(), gs.anim, gs.world)
	gs.action = action.New(gs.world, gs.mapgen, gs.query, gs.fx)

	gs.world.Player = mob.NewPC(gs.world, &data.PcSpec)
	gs.action.NextLevel()
}

func (gs *game) Exit() {}

func (gs *game) Draw() {
	sdl.Frame().Clear(gfx.Black)
	gfx.GradientRect(sdl.Frame(), image.Rect(0, 0, 320, 160), gfx.Green, gfx.ScaleCol(gfx.Green, 0.2))
	gs.view.Draw(image.Rect(4, 4, 316, 156))
	gs.hud.Draw(image.Rect(0, 0, 320, 240))
}

func (gs *game) Update(timeElapsed int64) {
	if gs.query.IsGameOver() {
		app.Get().PopState()
	}

	// Convenience maps for the directional keys.

	moveKeys := map[sdl.KeySym]image.Point{
		sdl.K_e: tile.HexDirs[0],
		sdl.K_r: tile.HexDirs[1],
		sdl.K_f: tile.HexDirs[2],
		sdl.K_d: tile.HexDirs[3],
		sdl.K_s: tile.HexDirs[4],
		sdl.K_w: tile.HexDirs[5]}

	shootKeys := map[sdl.KeySym]image.Point{
		sdl.K_i: tile.HexDirs[0],
		sdl.K_o: tile.HexDirs[1],
		sdl.K_l: tile.HexDirs[2],
		sdl.K_k: tile.HexDirs[3],
		sdl.K_j: tile.HexDirs[4],
		sdl.K_u: tile.HexDirs[5]}

	pc := gs.world.Player
	select {
	case evt := <-sdl.Events:
		switch e := evt.(type) {
		case sdl.KeyEvent:
			if e.KeyDown {
				if e.Sym == sdl.K_ESCAPE {
					app.Get().PopState()
					break
				}

				if dir, ok := moveKeys[e.FixedSym()]; ok {
					gs.action.AttackMove(pc, dir)
					gs.action.EndTurn()
					break
				}

				if dir, ok := shootKeys[e.FixedSym()]; ok {
					gs.action.Shoot(pc, dir)
					gs.action.EndTurn()
					break
				}

				// Layout independent keys
				switch e.FixedSym() {
				case sdl.K_SPACE:
					gs.action.EndTurn()
				case sdl.K_b:
					gs.fx.Blast(gs.query.Loc(pc), fx.SmallExplodeBlast)
					gs.action.Damage(gs.world.Player, 1)
				case sdl.K_n:
					gs.fx.Blast(gs.query.Loc(pc), fx.LargeExplodeBlast)
				}
			}
		case sdl.QuitEvent:
			app.Get().Stop()
		}
	default:
	}
}
