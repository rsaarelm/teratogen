package sdl

import (
	"time"
)

type Screen interface {
	Draw(intervalNs int64)
	Update(intervalNs int64)
}

func StartLoop(width, height int, title string, fullscreen bool) {
	InitSdl(width, height, title, fullscreen)
	running = true
	go mainLoop()
}

func StopLoop() {
	running = false
}

func mainLoop() {
	t := time.Nanoseconds()
	for running {
		currentT := time.Nanoseconds()
		if currentT - t > delayNs {
			// TODO Update screen..
			t = currentT
		}
		// TODO Sleep to avoid busy waiting.

		// TODO Draw screen if needed.

		// TODO Event handling with channels.
	}

	ExitSdl()
}

var screen Screen
var running bool
var delayNs int64 = 30 * 1e6
