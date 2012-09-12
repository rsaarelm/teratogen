// intro.go
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
	"fmt"
	"image"
	"math/rand"
	"teratogen/app"
	"teratogen/data"
	"teratogen/font"
	"teratogen/gfx"
	"teratogen/sdl"
)

func Intro() (in *intro) {
	in = new(intro)
	in.pcSelect = rand.Intn(data.NumClasses())
	return
}

type intro struct {
	pcSelect int
}

func (in *intro) Enter() {}
func (in *intro) Exit()  {}

func (in *intro) Draw() {
	sdl.Frame().Clear(gfx.Black)
	f, err := app.Cache().GetFont(font.Spec{"assets/BMmini.ttf", 8.0, 32, 96})
	if err != nil {
		panic(err)
	}
	cur := &font.Cursor{f, sdl.Frame(), image.Pt(0, 10), font.None, gfx.Green, gfx.Black}
	fmt.Fprintf(cur, "TERATOGEN")

	app.Cache().GetDrawable(data.PcPortrait[in.pcSelect]).Draw(image.Pt(0, 216))

	cur.Pos = image.Pt(24, 224)
	fmt.Fprintf(cur, data.PcDescr[in.pcSelect])
}

func (in *intro) Update(timeElapsed int64) {
	select {
	case evt := <-sdl.Events:
		switch e := evt.(type) {
		case sdl.KeyEvent:
			if e.KeyDown {
				if e.Sym == sdl.K_ESCAPE {
					app.Get().PopState()
				} else {
					switch e.FixedSym() {
					case sdl.K_n, sdl.K_RETURN, sdl.K_SPACE, sdl.K_KP_ENTER:
						app.Get().PushState(Game(in.pcSelect))
					case sdl.K_q, sdl.K_a, sdl.K_LEFT, sdl.K_KP4:
						in.pcSelect = (in.pcSelect + data.NumClasses() - 1) % data.NumClasses()
					case sdl.K_e, sdl.K_d, sdl.K_RIGHT, sdl.K_KP6:
						in.pcSelect = (in.pcSelect + 1) % data.NumClasses()
					}
				}
			}
		case sdl.QuitEvent:
			app.Get().Stop()
		}
	default:
	}
}
