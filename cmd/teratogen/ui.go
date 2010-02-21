package main

import (
	"container/vector"
	"exp/draw"
	"exp/iterable"
	"fmt"
	"hyades/alg"
	"hyades/dbg"
	"hyades/entity"
	"hyades/gfx"
	"hyades/gui"
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
const baseScreenWidth = 768
const baseScreenHeight = 480

var screenWidth = baseScreenWidth * 2
var screenHeight = baseScreenHeight * 2

const numFont = 256

const xDrawOffset = 0
const yDrawOffset = 0

const baseFontW = 8
const baseFontH = 8

const baseTileW = 8
const baseTileH = 8

var FontW = baseFontW * 2
var FontH = baseFontH * 2

var TileW = baseTileW * 2
var TileH = baseTileH * 2

type UI struct {
	gui.KeyHandlerStack

	gfx.Anims
	context sdl.Context
	msg     *MsgOut
	running bool

	// Show message lines beyond this to player.
	oldestLineSeen int

	mapView   *MapView
	timePoint int64
}

var ui *UI

var uiMutex = new(sync.Mutex)

var keymap = keyboard.KeyMap(keyboard.QwertyMap)

func GetUISync() { uiMutex.Lock() }

func ReleaseUISync() { uiMutex.Unlock() }

func newUI() (result *UI) {
	result = new(UI)
	result.KeyHandlerStack.Init()
	result.InitAnims()
	context, err := sdl.NewWindow(sdl.Config{
		Width: screenWidth, Height: screenHeight,
		Title: "Teratogen", Fullscreen: config.Fullscreen,
		Audio: config.Sound,
	})

	dbg.AssertNoError(err)
	result.context = context
	context.KeyRepeatOn()
	result.msg = NewMsgOut()
	result.running = true
	result.mapView = NewMapView()
	result.timePoint = time.Nanoseconds()
	result.PushKeyHandler(result.mapView)

	return
}

func (self *UI) AddMapAnim(anim *gfx.Anim) *gfx.Anim {
	return self.mapView.AddAnim(anim)
}

func (self *UI) AddScreenAnim(anim *gfx.Anim) *gfx.Anim {
	return self.AddAnim(anim)
}

func (self *UI) Draw(g gfx.Graphics, area draw.Rectangle) {
	g.FillRect(area, image.RGBAColor{0, 0, 0, 255})

	gui.DrawChildren(g, area, self)

	elapsed := time.Nanoseconds() - self.timePoint
	self.timePoint += elapsed

	self.DrawAnims(g, elapsed)
}

func (self *UI) Children(area draw.Rectangle) iterable.Iterable {
	// TODO: Adapt to area.
	cols, rows := 480/TileW-1, 320/TileH-1
	return alg.IterFunc(func(c chan<- interface{}) {
		c <- &gui.Window{self.mapView, draw.Rect(0, 0, TileW*cols, TileH*rows)}
		c <- &gui.Window{gui.DrawFunc(drawMsgLines),
			draw.Rect(0, TileH*(rows+1), screenWidth, screenHeight)}
		c <- &gui.Window{gui.DrawFunc(drawStatus),
			draw.Rect(TileW*(cols+1), 0, screenWidth, FontH*20)}
		close(c)
	})
}

func InitUI() { ui = newUI() }

func DrawSprite(g gfx.Graphics, name string, x, y int) {
	g.Blit(Media(name).(image.Image), x, y)
}

func DrawChar(g gfx.Graphics, char int, x, y int) {
	// XXX: Ineffctive string composition...
	if char > numFont {
		return
	}
	DrawSprite(g, fmt.Sprintf("font:%d", char), x, y)
}

// TODO: Support color
func DrawString(g gfx.Graphics, x, y int, format string, a ...interface{}) {
	for _, char := range fmt.Sprintf(format, a) {
		DrawChar(g, char, x, y)
		x += FontW
	}
}

func GetMsg() *MsgOut { return ui.msg }

func Msg(format string, a ...interface{}) { fmt.Fprintf(ui.msg, format, a) }

func GameRunning() bool { return ui.running }

func Quit() { ui.running = false }

func MarkMsgLinesSeen() { ui.oldestLineSeen = ui.msg.NumLines() - 1 }

// Blocking getkey function to be called from within an UI-locking game
// script. Unlocks the UI while waiting for key.
func GetKey() (result int) {
	ret := make(chan int)
	ui.PushKeyHandler(gui.KeyHandlerFunc(func(keyCode int) { ret <- keyCode }))
	defer ui.PopKeyHandler()

	ReleaseUISync()
	result = <-ret
	GetUISync()

	return
}

// Print --more-- and wait until the user presses space until proceeding.
func MsgMore() {
	Msg("--more--")
	for GetKey() != ' ' {
	}
}

func drawMsgLines(g gfx.Graphics, area draw.Rectangle) {
	g.SetClip(area)
	defer g.ClearClip()

	for i := ui.oldestLineSeen; i < GetMsg().NumLines(); i++ {
		DrawString(g, area.Min.X, area.Min.Y+(FontH*(i-ui.oldestLineSeen)),
			GetMsg().GetLine(i))
	}
}

func drawStatus(g gfx.Graphics, area draw.Rectangle) {
	g.SetClip(area)
	defer g.ClearClip()

	DrawString(g, area.Min.X, area.Min.Y,
		"%v", txt.Capitalize(GetCreature(PlayerId()).WoundDescription()))

	helpLineY := FontH * 3
	for o := range UiHelpLines().Iter() {
		DrawString(g, area.Min.X, area.Min.Y+helpLineY, o.(string))
		helpLineY += FontH
	}

}

func MainUILoop() {
	updater := time.Tick(redrawIntervalNs)
	lastTime := time.Nanoseconds()
	timeElapsed := int64(0)

	var prevMouseReceiver gui.MouseListener = nil

	for ui.running {
		if capFps {
			// Wait for the next tick before repainting.
			<-updater
		}
		timeElapsed = time.Nanoseconds() - lastTime
		lastTime += timeElapsed

		// Synched block which accesses the game world. Don't run
		// scripts during this.
		GetUISync()

		g := ui.context.SdlScreen()
		area := draw.Rect(0, 0, g.Width(), g.Height())
		ui.Draw(g, area)

		gui.DispatchTickEvent(ui, area, timeElapsed)

		if mouseEvt, ok := <-ui.context.MouseChan(); ok {
			prevMouseReceiver = gui.DispatchMouseEvent(area, ui, mouseEvt, prevMouseReceiver)
		}

		if keyEvt, ok := <-ui.context.KeyboardChan(); ok {
			ui.PeekKeyHandler().HandleKey(keyEvt)
		}

		if _, ok := <-ui.context.QuitChan(); ok {
			Quit()
		}

		ReleaseUISync()

		ui.context.FlushImage()
	}
	ui.context.Close()
}

func MultiChoiceDialog(prompt string, options ...) (choice int, ok bool) {
	return MultiChoiceDialogA(prompt, alg.UnpackEllipsis(options))
}

func MultiChoiceDialogA(prompt string, options []interface{}) (choice int, ok bool) {
	if len(options) == 0 {
		// Automatic abort on empty list.
		return -1, false
	}

	// TODO: More structured positioning.
	numVisible := 10
	xOff := 0
	yOff := TileH * 21
	lineH := FontH
	MarkMsgLinesSeen()
	pos := 0

	// Set running to false to shut off the animation for the dialog.
	running := true
	defer func() { running = false }()

	// Display function.
	go func(anim *gfx.Anim) {
		defer anim.Close()
		for running {
			g, _ := anim.StartDraw()
			moreAbove := pos > 0
			moreBelow := len(options)-pos > numVisible

			DrawString(g, xOff, yOff, prompt)
			if moreAbove {
				DrawString(g, xOff, yOff+lineH, "--more--")
			}
			for i := pos; i < num.Imin(pos+numVisible, len(options)); i++ {
				key := i - pos + 1
				if key == 10 {
					key = 0
				}
				DrawString(g, xOff, yOff+(2+i-pos)*lineH, "%d) %v", key, options[i])
			}
			if moreBelow {
				DrawString(g, xOff, yOff+(numVisible+2)*lineH, "--more--")
			}

			anim.StopDraw()
		}
	}(ui.AddScreenAnim(gfx.NewAnim(0.0)))

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
			if choice < maxOpt {
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
	idx, ok := MultiChoiceDialogA(prompt, names)
	if ok {
		result = objs[idx]
	}
	return
}

func EquipMenu() {
	player := GetBlob(PlayerId())
	slots := [...]string{PropBodyArmorGuid, PropMeleeWeaponGuid, PropGunWeaponGuid}
	names := [...]string{"body armor", "melee weapon", "gun"}
	options := make([]interface{}, len(slots))
	items := make([]interface{}, len(slots))
	for i, prop := range slots {
		if id, ok := GetEquipment(player.GetGuid(), prop); ok {
			ent := GetBlobs().Get(id).(*Blob)
			items[i] = ent
			options[i] = fmt.Sprintf("%s: %v", names[i], ent)
		} else {
			options[i] = fmt.Sprintf("%s: <nothing>", names[i])
		}
	}

	choice, ok := MultiChoiceDialogA("Equip/unequip item", options)
	if !ok {
		Msg("Okay, then.\n")
		return
	}
	if items[choice] != nil {
		RemoveEquipment(player.GetGuid(), slots[choice])
		Msg("Unequipped %v.\n", items[choice])
	} else {
		equippables := iterable.Data(iterable.Filter(iterable.Map(Contents(player.GetGuid()), id2Blob),
			func(o interface{}) bool { return CanEquipIn(slots[choice], o.(*Blob)) }))
		if item, ok := ObjectChoiceDialog(fmt.Sprintf("Equip %s", names[choice]), equippables); ok {
			SetEquipment(player.GetGuid(), slots[choice], item.(*Blob).GetGuid())
			Msg("Equipped %v.\n", item)
		}
	}
}

func UiHelpLines() iterable.Iterable {
	vec := new(vector.Vector)
	vec.Push("esc: exit menu")
	vec.Push("arrow keys: move, attack adjacent")
	vec.Push("q: quit")

	player := GetBlob(PlayerId())

	if HasContents(player.GetGuid()) {
		vec.Push("i: inventory")
		vec.Push("d: drop item")
	}

	if HasUsableItems(player) {
		vec.Push("a: use item")
	}

	if IsCarryingGear(player) {
		vec.Push("e: equip/remove gear")
	}

	if GunEquipped(player) {
		vec.Push("f: fire gun")
	}

	if len(iterable.Data(TakeableItems(GetPos(player.GetGuid())))) > 0 {
		vec.Push(",: pick up item")
	}
	if GetArea().GetTerrain(GetPos(player.GetGuid())) == TerrainStairDown {
		vec.Push(">: go down the stairs")
	}
	return vec
}

// Write a message about interesting stuff on the ground.
func StuffOnGroundMsg() {
	player := GetBlob(PlayerId())
	items := iterable.Data(TakeableItems(GetPos(player.GetGuid())))
	stairs := GetArea().GetTerrain(GetPos(player.GetGuid())) == TerrainStairDown
	if len(items) > 1 {
		Msg("There are several items here.\n")
	} else if len(items) == 1 {
		Msg("There is %v here.\n", GetBlob(items[0].(entity.Id)))
	}
	if stairs {
		Msg("There are stairs down here.\n")
	}
}

func ApplyItemMenu() (actionMade bool) {
	player := GetBlob(PlayerId())

	items := iterable.Data(iterable.Map(iterable.Filter(Contents(PlayerId()),
		EntityFilterFn(IsUsable)),
		id2Blob))
	if len(items) == 0 {
		Msg("You have no usable items.\n")
		return false
	}
	if item, ok := ObjectChoiceDialog("Use which item?", items); ok {
		UseItem(player, item.(*Blob))
		return true
	} else {
		Msg("Okay, then.\n")
	}
	return false
}
