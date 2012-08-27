// app.go
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

package app

import (
	"teratogen/gfx"
	"teratogen/sdl"
	"time"
)

// App is the toplevel object of an interactive game application. It handles
// different application state objects, maintaining the framerate.
type App interface {
	// Run runs the App until the app has no AppStates.
	Run()
	Stop()

	TopState() AppState
	PushState(as AppState)
	PopState()
}

type app struct {
	nanosecondsPerFrame int64
	states              []AppState
}

func (a *app) Run() {
	// If updates start
	const maxMultipleUpdates = 6

	lastTime := time.Now().UnixNano()

	for len(a.states) > 0 {
		currentTime := time.Now().UnixNano()
		if currentTime-lastTime < a.nanosecondsPerFrame {
			// Avoid busy waiting and take short naps if ahead of schedule.
			// XXX: Is this a good thing?
			time.Sleep(10e6)
			continue
		}

		// If things get slow, tell Update multiple frames have elapsed.
		nUpdates := (currentTime - lastTime) / a.nanosecondsPerFrame

		if nUpdates < 1 {
			nUpdates = 1
		}

		if nUpdates > maxMultipleUpdates {
			nUpdates = maxMultipleUpdates
		}

		a.TopState().Draw()
		gfx.BlitX3(sdl.Frame(), sdl.Video())
		sdl.Flip()

		a.TopState().Update(nUpdates * a.nanosecondsPerFrame)
		lastTime += nUpdates * a.nanosecondsPerFrame
	}
	sdl.Stop()
}

func (a *app) Stop() {
	for len(a.states) > 0 {
		a.PopState()
	}
}

func (a *app) TopState() AppState {
	if len(a.states) == 0 {
		return nil
	}
	return a.states[len(a.states)-1]
}

func (a *app) PushState(as AppState) {
	a.states = append(a.states, as)
	as.Enter()
}

func (a *app) PopState() {
	if len(a.states) > 0 {
		a.TopState().Exit()
		a.states = a.states[:len(a.states)-1]
	}
}

var globalApp App = nil

func Get() App {
	if globalApp == nil {
		globalApp = initApp()
	}
	return globalApp
}

func initApp() App {
	sdl.Run(960, 720)
	sdl.SetFrame(sdl.NewSurface(320, 240))

	a := &app{}
	a.nanosecondsPerFrame = 33e6
	a.states = []AppState{}

	return a
}

type AppState interface {
	Enter()
	Exit()
	Draw()
	Update(timeElapsed int64)
}
