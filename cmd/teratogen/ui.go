package main

import (
	"container/vector"
	"exp/draw"
	"fmt"
	"hyades/alg"
	"hyades/dbg"
	"hyades/sdl"
	"hyades/txt"
	"image"
	"sync"
	"time"
)

const redrawIntervalNs = 30e6
const capFps = true

// 16:10 aspect ratio
const screenWidth = 768 * 2
const screenHeight = 480 * 2

const numFont = 256

const xDrawOffset = 0
const yDrawOffset = 0

const TileW = 16
const TileH = 16

type UI struct {
	context sdl.Context
	msg     *MsgOut
	running bool

	// Show message lines beyond this to player.
	oldestLineSeen int

	anims *vector.Vector
}

var ui *UI

var uiMutex = new(sync.Mutex)

func GetUISync() { uiMutex.Lock() }

func ReleaseUISync() { uiMutex.Unlock() }

func newUI() (result *UI) {
	result = new(UI)
	context, err := sdl.NewWindow(screenWidth, screenHeight, "Teratogen", false)
	dbg.AssertNoError(err)
	result.context = context
	context.KeyRepeatOn()
	result.msg = NewMsgOut()
	result.running = true
	result.anims = new(vector.Vector)

	return
}

func animSort(i, j interface{}) bool { return i.(*Anim).Z < j.(*Anim).Z }

func AnimTest() { go TestAnim(ui.context, ui.AddAnim(NewAnim(0.0))) }

func (self *UI) AddAnim(anim *Anim) *Anim {
	self.anims.Push(anim)
	return anim
}

func (self *UI) DrawAnims(timeElapsedNs int64) {
	alg.PredicateSort(animSort, self.anims)
	for i := 0; i < self.anims.Len(); i++ {
		anim := self.anims.At(i).(*Anim)
		if anim.Closed() {
			self.anims.Delete(i)
			i--
			continue
		}
		// Tell the anim it can draw itself.
		anim.UpdateChan <- timeElapsedNs
		// Wait for the anim to call back that it's completed drawing itself.
		<-anim.UpdateChan
	}
}

func InitUI() { ui = newUI() }

func DrawSprite(name string, x, y int) { ui.context.Blit(Media(name).(image.Image), x, y) }

func DrawChar(char int, x, y int) {
	// XXX: Ineffctive string composition...
	if char > numFont {
		return
	}
	DrawSprite(fmt.Sprintf("font:%d", char), x, y)
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
func GetKey() (result int) {
	ReleaseUISync()
	for {
		result = <-ui.context.KeyboardChan()
		if result > 0 {
			break
		}
	}
	GetUISync()
	return
}

// Print --more-- and wait until the user presses space until proceeding.
func MsgMore() {
	Msg("--more--")
	for GetKey() != ' ' {
	}
}

func MainUILoop() {
	updater := time.Tick(redrawIntervalNs)
	lastTime := time.Nanoseconds()
	timeElapsed := int64(0)

	for ui.running {
		if capFps {
			// Wait for the next tick before repainting.
			<-updater
		}
		timeElapsed = time.Nanoseconds() - lastTime
		lastTime += timeElapsed

		ui.context.FillRect(draw.Rect(0, 0, screenWidth, screenHeight),
			image.RGBAColor{0, 0, 0, 255})

		// Synched block which accesses the game world. Don't run
		// scripts during this.
		GetUISync()

		world := GetWorld()

		world.Draw()

		for i := ui.oldestLineSeen; i < GetMsg().NumLines(); i++ {
			DrawString(GetMsg().GetLine(i), TileW*0, TileH*(21+(i-ui.oldestLineSeen)))
		}
		DrawString(fmt.Sprintf("Strength: %v", txt.Capitalize(LevelDescription(world.GetPlayer().GetI(PropStrength)))),
			TileW*41, TileH*0)
		DrawString(fmt.Sprintf("%v", txt.Capitalize(world.GetPlayer().WoundDescription())),
			TileW*41, TileH*1)

		ui.DrawAnims(timeElapsed)

		ReleaseUISync()

		ui.context.FlushImage()
	}
	ui.context.Close()
}
