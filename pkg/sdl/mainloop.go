package sdl

import (
	"hyades/dbg"
	"hyades/event"
	"time"
)

type Screen interface {
	Draw(intervalNs int64)
	Update(intervalNs int64)
}

func StartLoop(width, height int, title string, fullscreen bool) {
	Init(width, height, title, fullscreen)
	running = true
	eventChan = make(chan event.Event)
	go eventListener(eventChan)
//	go mainLoop()
}

func StopLoop() {
	running = false
	Exit()
}

func Events() <-chan event.Event { return eventChan }

// Sets the FPS cap. The WaitFrame function will sleep if it's called faster
// than the FPS interval.
func SetMaxFps(fps float64) {
	dbg.Assert(fps > 0, "Bad FPS")
	delayNs = int64(1e9 / fps)
	ticker = time.NewTicker(delayNs)
}

func WaitFrame() {
	<-ticker.C
}

func eventListener(ch chan event.Event) {
	for running {
		ch <- WaitEvent()
	}
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

	Exit()
}

var screen Screen
var running bool
var delayNs int64 = 30 * 1e6
var eventChan chan event.Event
var ticker *time.Ticker = time.NewTicker(delayNs)
