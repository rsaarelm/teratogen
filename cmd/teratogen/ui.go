package main

import (
	"container/vector"
	"exp/draw"
	"exp/iterable"
	"fmt"
	"hyades/alg"
	"hyades/dbg"
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
const screenWidth = 768 * 2
const screenHeight = 480 * 2

const numFont = 256

const xDrawOffset = 0
const yDrawOffset = 0

const FontH = 16
const FontW = 16

type UI struct {
	context sdl.Context
	msg     *MsgOut
	running bool

	// Show message lines beyond this to player.
	oldestLineSeen int

	mapView *MapView
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
	result.mapView = NewMapView()

	return
}

func (self *UI) AddAnim(anim *gfx.Anim) *gfx.Anim {
	return self.mapView.AddAnim(anim)
}

func (self *UI) Draw(g gfx.Graphics, area draw.Rectangle) {
	g.FillRect(area, image.RGBAColor{0, 0, 0, 255})

	gui.DrawChildren(g, area, self)
}

func (self *UI) Children(area draw.Rectangle) iterable.Iterable {
	// TODO: Adapt to area.
	// TODO: Widgetify main world view.
	return alg.IterFunc(func(c chan<- interface{}) {
		c <- gui.PackWidgetIteration(draw.Rect(0, 0, TileW*40, TileH*20),
			self.mapView)
		c <- gui.PackWidgetIteration(draw.Rect(0, TileH*21, screenWidth, screenHeight),
			gui.DrawFunc(drawMsgLines))
		c <- gui.PackWidgetIteration(draw.Rect(TileW*41, 0, screenWidth, FontH*20),
			gui.DrawFunc(drawStatus))
		close(c)
	})
}

func (self *UI) HandleMouseEvent(area draw.Rectangle, event draw.Mouse) {
	// TODO: Mouse response.
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
func DrawString(g gfx.Graphics, x, y int, format string, a ...) {
	for _, char := range fmt.Sprintf(format, a) {
		DrawChar(g, char, x, y)
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

func drawMsgLines(g gfx.Graphics, area draw.Rectangle) {
	//	g.SetClipRect(area)
	//	defer g.ClearClipRect()

	for i := ui.oldestLineSeen; i < GetMsg().NumLines(); i++ {
		DrawString(g, area.Min.X, area.Min.Y+(FontH*(i-ui.oldestLineSeen)),
			GetMsg().GetLine(i))
	}
}

func drawStatus(g gfx.Graphics, area draw.Rectangle) {
	//	g.SetClipRect(area)
	//	defer g.ClearClipRect()

	DrawString(g, area.Min.X, area.Min.Y,
		"%v", txt.Capitalize(world.GetPlayer().WoundDescription()))

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
		ui.Draw(g, draw.Rect(0, 0, g.Width(), g.Height()))

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
	lineH := TileH
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
			for i := pos; i < num.IntMin(pos+numVisible, len(options)); i++ {
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
	}(ui.AddAnim(gfx.NewAnim(0.0)))

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
	player := GetWorld().GetPlayer()
	slots := [...]string{PropBodyArmorGuid, PropMeleeWeaponGuid, PropGunWeaponGuid}
	names := [...]string{"body armor", "melee weapon", "gun"}
	options := make([]interface{}, len(slots))
	items := make([]interface{}, len(slots))
	for i, prop := range slots {
		if ent, ok := player.GetGuidOpt(prop); ok {
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
		player.Clear(slots[choice])
		Msg("Unequipped %v.\n", items[choice])
	} else {
		equippables := iterable.Data(iterable.Filter(player.Contents(),
			func(o interface{}) bool { return CanEquipIn(slots[choice], o.(*Entity)) }))
		if item, ok := ObjectChoiceDialog(fmt.Sprintf("Equip %s", names[choice]), equippables); ok {
			player.Set(slots[choice], item.(*Entity).GetGuid())
			Msg("Equipped %v.\n", item)
		}
	}
}

func UiHelpLines() iterable.Iterable {
	vec := new(vector.Vector)
	vec.Push("esc: exit menu")
	vec.Push("arrow keys: move, attack adjacent")
	vec.Push("q: quit")

	world := GetWorld()
	player := world.GetPlayer()

	if HasContents(player) {
		vec.Push("i: inventory")
		vec.Push("d: drop item")
	}

	if HasUsableItems(player) {
		vec.Push("a: use item")
	}

	if IsCarryingGear(player) {
		vec.Push("e: equip/remove gear")
	}

	if len(iterable.Data(TakeableItems(player.GetPos()))) > 0 {
		vec.Push(",: pick up item")
	}
	if world.GetTerrain(player.GetPos()) == TerrainStairDown {
		vec.Push(">: go down the stairs")
	}
	return vec
}

// Write a message about interesting stuff on the ground.
func StuffOnGroundMsg() {
	world := GetWorld()
	player := world.GetPlayer()
	items := iterable.Data(TakeableItems(player.GetPos()))
	stairs := world.GetTerrain(player.GetPos()) == TerrainStairDown
	if len(items) > 1 {
		Msg("There are several items here.\n")
	} else if len(items) == 1 {
		Msg("There is %v here.\n", items[0])
	}
	if stairs {
		Msg("There are stairs down here.\n")
	}
}

func ApplyItemMenu() (actionMade bool) {
	player := GetWorld().GetPlayer()

	items := iterable.Data(iterable.Filter(player.Contents(), IsUsable))
	if len(items) == 0 {
		Msg("You have no usable items.\n")
		return false
	}
	if item, ok := ObjectChoiceDialog("Use which item?", items); ok {
		UseItem(player, item.(*Entity))
		return true
	} else {
		Msg("Okay, then.\n")
	}
	return false
}
