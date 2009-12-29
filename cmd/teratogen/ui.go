package main

import (
	"container/vector"
	"exp/draw"
	"fmt"
	"hyades/alg"
	"hyades/dbg"
	"hyades/keyboard"
	"hyades/num"
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

// TODO: Configure externally.
var keymap = keyboard.KeyMap(keyboard.ColemakMap)

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
func DrawString(x, y int, format string, a ...) {
	for _, char := range fmt.Sprintf(format, a) {
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
			DrawString(TileW*0, TileH*(21+(i-ui.oldestLineSeen)), GetMsg().GetLine(i))
		}
		DrawString(TileW*41, TileH*0,
			"Strength: %v", txt.Capitalize(LevelDescription(world.GetPlayer().GetI(PropStrength))))
		DrawString(TileW*41, TileH*1,
			"%v", txt.Capitalize(world.GetPlayer().WoundDescription()))

		ui.DrawAnims(timeElapsed)

		ReleaseUISync()

		ui.context.FlushImage()
	}
	ui.context.Close()
}

func MultiChoiceDialog(prompt string, options ...) (choice int, ok bool) {
	return MultiChoiceDialogV(prompt, alg.UnpackEllipsis(options))
}

func MultiChoiceDialogV(prompt string, options []interface{}) (choice int, ok bool) {
	// TODO: More structured positioning.
	numVisible := 10
	xOff := 0
	yOff := TileH * 21
	lineH := TileH
	MarkMsgLinesSeen()
	pos := 0

	// Set running to false to shut off the animation for the dialog.
	running := true
	defer func() { running = false }()

	// Display function.
	go func(anim *Anim) {
		defer anim.Close()
		for running {
			<-anim.UpdateChan
			moreAbove := pos > 0
			moreBelow := len(options)-pos > numVisible

			DrawString(xOff, yOff, prompt)
			if moreAbove {
				DrawString(xOff, yOff+lineH, "--more--")
			}
			for i := pos; i < num.IntMin(pos+numVisible, len(options)); i++ {
				key := i - pos + 1
				if key == 10 {
					key = 0
				}
				DrawString(xOff, yOff+(2+i-pos)*lineH, "%d) %v", key, options[i])
			}
			if moreBelow {
				DrawString(xOff, yOff+(numVisible+2)*lineH, "--more--")
			}

			anim.UpdateChan <- 0
		}
	}(ui.AddAnim(NewAnim(0.0)))

	for {
		moreAbove := pos > 0
		moreBelow := len(options)-pos > numVisible
		maxOpt := len(options) - pos

		key := keymap.Map(GetKey())
		switch {
		case key == 'k' || key == keyboard.K_UP || key == keyboard.K_KP8 || key == keyboard.K_PAGEUP:
			if moreAbove {
				pos -= numVisible
			}
		case key == 'j' || key == keyboard.K_DOWN || key == keyboard.K_KP2 || key == keyboard.K_PAGEDOWN:
			if moreBelow {
				pos += numVisible
			}
		// TODO: PgUp, PgDown
		case key >= '0' && key <= '9':
			choice := key - '1'
			if choice == -1 {
				// Correct for the position of ASCII '0'
				choice += 10
			}
			if choice <= maxOpt {
				return choice + pos, true
			}
		case key == keyboard.K_ESCAPE:
			// Return index -1 along with not-ok if the user
			// aborts, so that buggy calling code that tries to
			// use the return value despite getting not-ok will
			// fail faster.
			return -1, false
		}
	}
	panic("MultiChoiceDialog exited unexpectedly")
}

func ObjectChoiceDialog(prompt string, objs []interface{}) (result interface{}, ok bool) {
	names := make([]interface{}, len(objs))
	for i, obj := range objs {
		if stringer, ok := obj.(fmt.Stringer); ok {
			names[i] = stringer.String()
		} else {
			names[i] = fmt.Sprint(obj)
		}
	}
	idx, ok := MultiChoiceDialogV(prompt, names)
	if ok {
		result = objs[idx]
	}
	return
}
