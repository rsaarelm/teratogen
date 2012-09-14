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
	"math/rand"
	"teratogen/action"
	"teratogen/app"
	"teratogen/data"
	"teratogen/display"
	"teratogen/gfx"
	"teratogen/manifold"
	"teratogen/mob"
	"teratogen/sdl"
	"teratogen/world"
)

func Game(pcSelect int) (gs *game) {
	gs = new(game)
	gs.pcSelect = pcSelect
	return
}

type game struct {
	world    *world.World
	action   *action.Action
	disp     *display.Display
	pcSelect int
}

func (gs *game) Enter() {
	gs.world = world.New()
	gs.action = action.New(gs.world)

	origin := manifold.Location{0, 0, 1}
	gs.world.TestMap(origin)

	bounds := image.Rect(-16, -16, 16, 16)
	for i := 0; i < 32; i++ {
		pos := image.Pt(rand.Intn(bounds.Dx())+bounds.Min.X, rand.Intn(bounds.Dy())+bounds.Min.Y)
		m := mob.New(gs.world, &mob.Spec{gfx.ImageSpec{"assets/chars.png", image.Rect(8, 0, 16, 8)}})
		loc := origin.Add(pos)
		if gs.action.Fits(m, loc) {
			gs.action.Place(m, loc)
		}
	}

	pc := mob.NewPC(gs.world, &data.PcSpec[gs.pcSelect])

found:
	for {
		for i := 0; i < 64; i++ {
			pos := image.Pt(rand.Intn(bounds.Dx())+bounds.Min.X, rand.Intn(bounds.Dy())+bounds.Min.Y)
			loc := origin.Add(pos)
			if gs.action.Fits(pc, loc) {
				gs.action.Place(pc, loc)
				gs.world.Player = pc
				break found
			}
		}
		panic("Can't place player")
	}

	gs.disp = display.New(app.Cache(), gs.world)
}

func (gs *game) Exit() {}

func (gs *game) Draw() {
	sdl.Frame().Clear(gfx.Black)
	gfx.GradientRect(sdl.Frame(), image.Rect(0, 0, 320, 160), gfx.Green, gfx.ScaleCol(gfx.Green, 0.2))
	gs.disp.DrawWorld(image.Rect(4, 4, 316, 156))
	gs.disp.DrawMsg(image.Rect(2, 162, 158, 238))

	app.Cache().GetDrawable(data.PcPortrait[gs.pcSelect]).Draw(image.Pt(0, 216))
}

func (gs *game) Update(timeElapsed int64) {
	if gs.action.IsGameOver() {
		app.Get().PopState()
	}

	pc := gs.world.Player
	select {
	case evt := <-sdl.Events:
		switch e := evt.(type) {
		case sdl.KeyEvent:
			if e.KeyDown {
				if e.Sym == sdl.K_ESCAPE {
					app.Get().PopState()
				}

				// Layout independent keys
				switch e.FixedSym() {
				case sdl.K_q:
					gs.action.AttackMove(pc, image.Pt(-1, 0))
					gs.action.EndTurn()
				case sdl.K_w:
					gs.action.AttackMove(pc, image.Pt(-1, -1))
					gs.action.EndTurn()
				case sdl.K_e:
					gs.action.AttackMove(pc, image.Pt(0, -1))
					gs.action.EndTurn()
				case sdl.K_d:
					gs.action.AttackMove(pc, image.Pt(1, 0))
					gs.action.EndTurn()
				case sdl.K_s:
					gs.action.AttackMove(pc, image.Pt(1, 1))
					gs.action.EndTurn()
				case sdl.K_a:
					gs.action.AttackMove(pc, image.Pt(0, 1))
					gs.action.EndTurn()
				case sdl.K_SPACE:
					gs.action.EndTurn()
				}
			}
		case sdl.QuitEvent:
			app.Get().Stop()
		}
	default:
	}
}
