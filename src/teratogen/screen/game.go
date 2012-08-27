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
	"teratogen/app"
	"teratogen/display"
	"teratogen/gfx"
	"teratogen/manifold"
	"teratogen/mob"
	"teratogen/sdl"
	"teratogen/world"
	"time"
)

func GameScreen() (gs *gameState) {
	return new(gameState)
}

type gameState struct {
	world *world.World
	disp  *display.Display
}

func (gs *gameState) Enter() {
	rand.Seed(time.Now().UnixNano())

	gs.world = world.New()

	origin := manifold.Location{0, 0, 1}
	gs.world.TestMap(origin)

	bounds := image.Rect(-16, -16, 16, 16)
	for i := 0; i < 32; i++ {
		pos := image.Pt(rand.Intn(bounds.Dx())+bounds.Min.X, rand.Intn(bounds.Dy())+bounds.Min.Y)
		m := mob.New(gs.world, &mob.Spec{gfx.ImageSpec{"assets/chars.png", image.Rect(8, 0, 16, 8)}})
		loc := origin.Add(pos)
		if m.Fits(loc) {
			m.Place(loc)
		}
	}

	pc := mob.New(gs.world, &mob.Spec{gfx.ImageSpec{"assets/chars.png", image.Rect(0, 16, 8, 24)}})

found:
	for {
		for i := 0; i < 64; i++ {
			pos := image.Pt(rand.Intn(bounds.Dx())+bounds.Min.X, rand.Intn(bounds.Dy())+bounds.Min.Y)
			loc := origin.Add(pos)
			if pc.Fits(loc) {
				pc.Place(loc)
				gs.world.Player = pc
				break found
			}
		}
		panic("Can't place player")
	}

	gs.disp = display.New(app.Cache(), gs.world)
}

func (gs *gameState) Exit() {}

func (gs *gameState) Draw() {
	gfx.GradientRect(sdl.Frame(), image.Rect(0, 0, 320, 160), gfx.Green, gfx.ScaleCol(gfx.Green, 0.2))
	gs.disp.DrawWorld(image.Rect(4, 4, 316, 156))
	gs.disp.DrawMsg(image.Rect(2, 162, 158, 238))

}

func (gs *gameState) Update(timeElapsed int64) {
	select {
	case evt := <-sdl.Events:
		switch e := evt.(type) {
		case sdl.KeyEvent:
			if e.KeyDown {
				if e.Sym == sdl.K_ESCAPE {
					app.Get().PopState()
					app.Get().PushState(IntroScreen())
				}

				// Layout independent keys
				switch e.FixedSym() {
				case sdl.K_q:
					gs.disp.Move(image.Pt(-1, 0))
				case sdl.K_w:
					gs.disp.Move(image.Pt(-1, -1))
				case sdl.K_e:
					gs.disp.Move(image.Pt(0, -1))
				case sdl.K_d:
					gs.disp.Move(image.Pt(1, 0))
				case sdl.K_s:
					gs.disp.Move(image.Pt(1, 1))
				case sdl.K_a:
					gs.disp.Move(image.Pt(0, 1))
				}
			}
		case sdl.QuitEvent:
			app.Get().Stop()
		}
	default:
	}
}
