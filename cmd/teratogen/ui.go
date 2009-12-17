package main

import (
	"fmt"
//	. "hyades/common"
	"hyades/event"
	"hyades/sdl"
//	"hyades/txt"
	"image"
	"sync"
	"time"
)

const redrawIntervalNs = 30e6
const capFps = true

const screenWidth = 640 * 2
const screenHeight = 480 * 2

type UI struct {
	msg	*MsgOut
	running	bool

	// Show message lines beyond this to player.
	oldestLineSeen	int
}

var ui *UI

var uiMutex = new(sync.Mutex)

func GetUISync()	{ uiMutex.Lock() }

func ReleaseUISync()	{ uiMutex.Unlock() }

func newUI() (result *UI) {
	result = new(UI)
	sdl.StartLoop(screenWidth, screenHeight, "Teratogen", false)
	sdl.KeyRepeatOn()
	result.msg = NewMsgOut()
	result.running = true

	return
}

func InitUI()	{ ui = newUI() }

func DrawSprite(name string, x, y int) {
	sprite := Media(name).(*sdl.Surface)
	sprite.Blit(sdl.GetVideoSurface(), x, y)
}

func GetMsg() *MsgOut	{ return ui.msg }

func Msg(format string, a ...)	{ fmt.Fprintf(ui.msg, format, a) }

func Quit()	{ ui.running = false }

func MarkMsgLinesSeen()	{ ui.oldestLineSeen = ui.msg.NumLines() - 1 }

// Blocking getkey function to be called from within an UI-locking game
// script. Unlocks the UI while waiting for key.
func GetKey() (result *event.KeyDown) {
	ReleaseUISync()
	for {
		switch evt := (<-sdl.Events()).(type) {
		case *event.KeyDown:
			return evt
		}
	}
	GetUISync()
	return
}

// Print --more-- and wait until the user presses space until proceeding.
func MsgMore() {
	Msg("--more--")
	for GetKey().Printable != ' ' {
	}
}

func MainUILoop() {
	updater := time.Tick(redrawIntervalNs)

	for ui.running {
		// XXX: Can't put grabbing and releasing sync next to each
		// other, to the very beginning and end of the loop, or the
		// script side will never get sync.

		if capFps {
			// Wait for the next tick before repainting.
			<-updater
		}

		sdl.GetVideoSurface().FillRect(
			sdl.Rect(0, 0, screenWidth, screenHeight),
			image.RGBAColor{0, 0, 0, 255})

		// Synched block which accesses the game world. Don't run
		// scripts during this.
		GetUISync()

		world := GetWorld()

		world.Draw()

//		for i := ui.oldestLineSeen; i < GetMsg().NumLines(); i++ {
//			con.Print(0, 21+(i-ui.oldestLineSeen), GetMsg().GetLine(i))
//		}
//
//		con.Print(41, 0, fmt.Sprintf("Strength: %v",
//			txt.Capitalize(LevelDescription(world.GetPlayer().Strength))))
//		con.Print(41, 1, fmt.Sprintf("%v",
//			txt.Capitalize(world.GetPlayer().WoundDescription())))
		ReleaseUISync()

		sdl.Flip()
	}
	sdl.StopLoop()
}
