package main

import (
	"fmt"
	"hyades/event"
	"hyades/sdl"
	"hyades/txt"
	"image"
	"sync"
	"time"
)

const redrawIntervalNs = 30e6
const capFps = true

const screenWidth = 640 * 2
const screenHeight = 480 * 2

const numFont = 384

type UI struct {
	msg     *MsgOut
	running bool

	// Show message lines beyond this to player.
	oldestLineSeen int
}

var ui *UI

var uiMutex = new(sync.Mutex)

var eventChan = make(chan event.Event, 16)

func PipeEvents() {
	for {
		evt := sdl.PollEvent()
		if evt != nil {
			roomLeft := eventChan <- evt
			if !roomLeft {
				// Drop old events
				_, _ = <-eventChan
				break
			}
		} else {
			break
		}
	}
}

func GetUISync() { uiMutex.Lock() }

func ReleaseUISync() { uiMutex.Unlock() }

func newUI() (result *UI) {
	result = new(UI)
	sdl.Init(screenWidth, screenHeight, "Teratogen", false)
	sdl.KeyRepeatOn()
	result.msg = NewMsgOut()
	result.running = true

	return
}

func InitUI() { ui = newUI() }

func DrawSprite(name string, x, y int) {
	sprite := Media(name).(*sdl.Surface)
	sprite.Blit(sdl.GetVideoSurface(), x, y)
}

func DrawChar(char int, x, y int) {
	// XXX: Ineffctive string composition...
	if char > numFont {
		return
	}
	Media(fmt.Sprintf("font:%d", char)).(*sdl.Surface).Blit(sdl.GetVideoSurface(), x, y)
}

// TODO: Support color
func DrawString(txt string, x, y int) {
	for _, char := range txt {
		DrawChar(char, x, y)
		x += TileW
	}
}

func GetMsg() *MsgOut { return ui.msg }

func Msg(format string, a ...) { fmt.Fprintf(ui.msg, format, a) }

func Quit() { ui.running = false }

func MarkMsgLinesSeen() { ui.oldestLineSeen = ui.msg.NumLines() - 1 }

// Blocking getkey function to be called from within an UI-locking game
// script. Unlocks the UI while waiting for key.
func GetKey() (result *event.KeyDown) {
	ReleaseUISync()
	for {
		switch evt := (<-eventChan).(type) {
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

		for i := ui.oldestLineSeen; i < GetMsg().NumLines(); i++ {
			DrawString(GetMsg().GetLine(i), TileW*0, TileH*(21+(i-ui.oldestLineSeen)))
		}
		DrawString(fmt.Sprintf("Strength: %v", txt.Capitalize(LevelDescription(world.GetPlayer().Strength))),
			TileW*41, TileH*0)
		DrawString(fmt.Sprintf("%v", txt.Capitalize(world.GetPlayer().WoundDescription())),
			TileW*41, TileH*1)

		PipeEvents()

		ReleaseUISync()

		sdl.Flip()
	}
	sdl.Exit()
}
