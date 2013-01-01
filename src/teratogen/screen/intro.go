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

// Package screen defines the toplevel application states for Teratogen.
package screen

import (
	"fmt"
	"image"
	"teratogen/app"
	"teratogen/font"
	"teratogen/gfx"
	"teratogen/sdl"
)

func Intro() app.State {
	in := new(intro)
	return in
}

type intro struct {
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
						app.Get().PushState(Game())
					}
				}
			}
		case sdl.QuitEvent:
			app.Get().Stop()
		}
	default:
	}
}
