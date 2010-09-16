package main

import (
	"container/vector"
	"exp/iterable"
	"fmt"
	"hyades/dbg"
	"hyades/entity"
	"hyades/geom"
	"hyades/gfx"
	"hyades/gui"
	"hyades/keyboard"
	"hyades/num"
	"hyades/sdl"
	"hyades/txt"
	"image"
	"sync"
	game "teratogen"
	"time"
)

const UseGzipSaves = true

const redrawIntervalNs = 30e6
const capFps = true

const baseScreenWidth = 640
const baseScreenHeight = 400

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
	font    sdl.Font
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
		Width:      screenWidth,
		Height:     screenHeight,
		PixelScale: config.Scale,
		Title:      "Teratogen",
		Fullscreen: config.Fullscreen,
		Audio:      config.Sound,
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

func (self *UI) Draw(g gfx.Graphics, area image.Rectangle) {
	g.FillRect(area, image.RGBAColor{0, 0, 0, 255})

	gui.DrawChildren(g, area, self)

	elapsed := time.Nanoseconds() - self.timePoint
	self.timePoint += elapsed

	self.DrawAnims(g, elapsed)
}

func (self *UI) Children(area image.Rectangle) iterable.Iterable {
	// TODO: Adapt to area.
	cols, rows := VisualScale()*200/TileW-1, VisualScale()*120/TileH-1
	return iterable.Func(func(c chan<- interface{}) {
		c <- &gui.Window{self.mapView, image.Rect(0, 0, TileW*cols, TileH*rows)}
		c <- &gui.Window{gui.DrawFunc(drawCommandBar),
			image.Rect(0, TileH*rows, screenWidth, TileH*(rows+2))}
		c <- &gui.Window{gui.DrawFunc(drawMsgLines),
			image.Rect(0, TileH*(rows+3), screenWidth, screenHeight)}
		c <- &gui.Window{gui.DrawFunc(drawStatus),
			image.Rect(TileW*(cols+1), 0, screenWidth, FontH*20)}
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

var defaultTextColor = gfx.GreenYellow

func DrawColorString(g gfx.Graphics, x, y int, col image.Color, format string, a ...interface{}) {
	txt := fmt.Sprintf(format, a)
	if txt == "" {
		return
	}
	line, err := ui.font.Render(txt, col)
	if err != nil {
		fmt.Printf("[%v], %v\n", txt, err)
		return
	}
	g.Blit(line, x, y)
	ui.context.Free(line)
}

func DrawString(g gfx.Graphics, x, y int, format string, a ...interface{}) {
	DrawColorString(g, x, y, defaultTextColor, format, a)
}

func GetMsg() *MsgOut { return ui.msg }

func GameRunning() bool { return ui.running }

func Quit() { ui.running = false }

func MarkMsgLinesSeen() { ui.oldestLineSeen = ui.msg.NumLines() - 1 }

// Blocking getkey function to be called from within an UI-locking game
// script. Unlocks the UI while waiting for key.
func GetKey() (result int) {
	ret := make(chan int)
	ui.PushKeyHandler(gui.KeyHandlerFunc(func(keyCode int) { ret <- keyCode }))
	defer ui.PopKeyHandler()

	result = -1

	// Don't return key release events, which have negative numbers as values.
	for result < 0 {
		ReleaseUISync()
		result = <-ret
		GetUISync()
	}

	return
}

func drawCommandBar(g gfx.Graphics, area image.Rectangle) {
	g.SetClip(area)
	defer g.ClearClip()
	const numSlots = 10
	const slotWidth = 16
	const perLine = 5

	for i := 0; i < numSlots; i++ {
		x := (i % perLine) * slotWidth * FontW
		y := (i / perLine) * FontH
		// TODO: Print text description
		idx := (i + 1) % 10
		txt := fmt.Sprintf("%d %s", idx, ShortNamePower(idx))
		DrawString(g, area.Min.X+x, area.Min.Y+y, txt)
	}
}

func drawMsgLines(g gfx.Graphics, area image.Rectangle) {
	g.SetClip(area)
	defer g.ClearClip()

	fontH := ui.font.Height()

	// Scroll down if there's too much text.
	maxLines := area.Dy() / fontH
	if GetMsg().NumLines()-ui.oldestLineSeen > maxLines {
		ui.oldestLineSeen = GetMsg().NumLines() - maxLines
	}

	for i := ui.oldestLineSeen; i < GetMsg().NumLines(); i++ {
		yOff := area.Min.Y + (fontH * (i - ui.oldestLineSeen))
		DrawString(g, area.Min.X, yOff,
			GetMsg().GetLine(i))
		if GetMsg().CursorPos.Y == i {
			// Draw cursor.
			xOff := ui.font.StringWidth(GetMsg().GetLine(i)[0:GetMsg().CursorPos.X])
			g.FillRect(
				image.Rect(area.Min.X+xOff, yOff,
					area.Min.X+xOff+1, yOff+ui.font.Height()),
				defaultTextColor)
		}
	}
}

func statusLineColor() image.Color {
	playerCrit := game.GetCreature(game.PlayerId())
	switch {
	case playerCrit.IsSeriouslyHurt():
		return gfx.OrangeRed
	case playerCrit.IsHurt():
		return gfx.Orange
	case playerCrit.HasStatus(game.StatusMutationShield):
		return gfx.LightBlue
	}
	return defaultTextColor
}

func drawStatus(g gfx.Graphics, area image.Rectangle) {
	g.SetClip(area)
	defer g.ClearClip()

	playerCrit := game.GetCreature(game.PlayerId())

	wounds := playerCrit.WoundDescription()

	healthStatus := wounds
	if playerCrit.Armor > 0 {
		healthStatus += fmt.Sprintf(" + %d Armor", int(playerCrit.Armor*game.ArmorScale))
	}
	DrawColorString(g, area.Min.X, area.Min.Y, statusLineColor(),
		"%v", txt.Capitalize(healthStatus))

	mutations := game.GetMutations(game.PlayerId())
	DrawColorString(g, area.Min.X, area.Min.Y+FontH, statusLineColor(),
		fmt.Sprintf("%.02f humanity", mutations.HumanityLevel()*100.0))

	inv := game.GetInventory(game.PlayerId())
	DrawString(g, area.Min.X, area.Min.Y+2*FontH, fmt.Sprintf("Ammo: %d", inv.Ammo))

	helpLineY := FontH * 4
	for _, o := range *UiHelpLines() {
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
		area := g.Bounds()
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

func MultiChoiceDialog(prompt string, options ...interface{}) (choice int, ok bool) {
	return MultiChoiceDialogA(prompt, options)
}

func MultiChoiceDialogA(prompt string, options []interface{}) (choice int, ok bool) {
	if len(options) == 0 {
		// Automatic abort on empty list.
		return -1, false
	}

	// TODO: More structured positioning.
	numVisible := 10
	xOff := 0
	yOff := TileH * 16
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
		case key == keyboard.K_ESCAPE || isActionKey(key):
			// Return index -1 along with not-ok if the user
			// aborts, so that buggy calling code that tries to
			// use the return value despite getting not-ok will
			// fail faster.
			return -1, false
		}
	}
	panic("MultiChoiceDialog exited unexpectedly")
}

func EntityChoiceDialog(prompt string, ids []interface{}) entity.Id {
	names := make([]interface{}, len(ids))
	for i, o := range ids {
		names[i] = game.GetName(o.(entity.Id))
	}
	idx, ok := MultiChoiceDialogA(prompt, names)
	if ok {
		return ids[idx].(entity.Id)
	}
	return entity.NilId
}

func EquipMenu() {
	subjectId := game.PlayerId()
	slots := [...]game.EquipSlot{game.ArmorEquipSlot, game.MeleeEquipSlot, game.GunEquipSlot}
	names := [...]string{"body armor", "melee weapon", "gun"}
	options := make([]interface{}, len(slots))
	items := make([]interface{}, len(slots))
	for i, prop := range slots {
		if id, ok := game.GetEquipment(subjectId, prop); ok {
			items[i] = id
			options[i] = fmt.Sprintf("%s: %v", names[i], game.GetName(id))
		} else {
			options[i] = fmt.Sprintf("%s: <nothing>", names[i])
		}
	}

	choice, ok := MultiChoiceDialogA("Equip/unequip item", options)
	if !ok {
		game.Msg("Okay, then.\n")
		return
	}
	if items[choice] != nil {
		game.RemoveEquipment(subjectId, slots[choice])
		game.Msg("Unequipped %v.\n", game.GetName(items[choice].(entity.Id)))
	}

	equippables := iterable.Data(iterable.Filter(game.Contents(subjectId),
		func(o interface{}) bool { return game.CanEquipIn(slots[choice], o.(entity.Id)) }))
	prompt := fmt.Sprintf("Equip %s", names[choice])
	if id := EntityChoiceDialog(prompt, equippables); id != entity.NilId {
		game.SetEquipment(subjectId, slots[choice], id)
		game.Msg("Equipped %v.\n", game.GetName(id))
	}
}

func UiHelpLines() *vector.Vector {
	vec := new(vector.Vector)
	vec.Push("esc: exit menu")
	vec.Push("arrow keys, qwe asd: move/attack")
	vec.Push("uio jkl: fire gun")
	vec.Push("Return, numpad 5: action key")
	vec.Push("Q: quit S: save and quit")
	vec.Push("Typewriter number keys: Command bar")

	if game.HasContents(game.PlayerId()) {
		vec.Push("t: inventory")
		vec.Push("D: drop item")
	}

	if game.HasUsableItems(game.PlayerId()) {
		vec.Push("A: use item")
	}

	if game.IsCarryingGear(game.PlayerId()) {
		vec.Push("E: equip/remove gear")
	}

	if game.GunEquipped(game.PlayerId()) {
		vec.Push("uiojkl: fire gun in direction")
	}

	if game.NumPlayerTakeableItems() > 0 {
		vec.Push(",: pick up item")
	}
	if game.PlayerAtStairs() {
		vec.Push(">: go down the stairs")
	}
	return vec
}

func isActionKey(keysym int) bool {
	return keysym == keyboard.K_RETURN || keysym == keyboard.K_KP5
}

func ApplyItemMenu() (actionMade bool) {
	items := iterable.Data(iterable.Filter(game.Contents(game.PlayerId()),
		game.EntityFilterFn(game.IsUsable)))
	if len(items) == 0 {
		game.Msg("You have no usable items.\n")
		return false
	}
	if id := EntityChoiceDialog("Use which item?", items); id != entity.NilId {
		SendPlayerInput(func() bool {
			game.UseItem(game.PlayerId(), id)
			return true
		})
		return true
	} else {
		game.Msg("Okay, then.\n")
	}
	return false
}

type SdlEffects struct{}

func (self *SdlEffects) Print(str string) { fmt.Fprint(ui.msg, str) }

func (self *SdlEffects) Shoot(shooterId entity.Id, hitPos geom.Pt2I, fx game.AttackFx) {
	worldP1 := Tile2WorldPos(game.GetPos(shooterId))
	worldP2 := Tile2WorldPos(hitPos).Plus(InCellJitter())
	p1, p2 := image.Pt(worldP1.X, worldP1.Y), image.Pt(worldP2.X, worldP2.Y)
	switch fx {
	case game.AttackFxBeam:
		go LineAnim(ui.AddMapAnim(gfx.NewAnim(0.0)), p1, p2, 2e8, gfx.White, gfx.DarkRed, VisualScale())
		//	case game.AttackFxSpray:
		//	case game.AttackFxElectro:
	default:
		go LineAnim(ui.AddMapAnim(gfx.NewAnim(0.0)), p1, p2, 2e8, gfx.White, gfx.DarkRed, VisualScale())
	}
}

var smallBlastEmit = ParticleBlastEmitter(2, 10)

func (self *SdlEffects) Sparks(pos geom.Pt2I) {
	p := Tile2WorldPos(pos).Plus(InCellJitter())
	go ParticleAnim(ui.AddMapAnim(gfx.NewAnim(0.0)), materialSpark,
		smallBlastEmit, p.X, p.Y, 15)
}

func (self *SdlEffects) Damage(id entity.Id, woundLevel int) {
	sx, sy := CenterDrawPos(game.GetPos(id))
	go ParticleAnim(ui.AddMapAnim(gfx.NewAnim(0.0)), materialBlood,
		smallBlastEmit, sx, sy, num.Imax(8, woundLevel))
	PlaySound("hit")
}

func (self *SdlEffects) Heal(id entity.Id, amount int) {
	PlaySound("heal")
}

func (self *SdlEffects) Destroy(id entity.Id) {
	sx, sy := CenterDrawPos(game.GetPos(id))
	const gibNum = 20
	go ParticleAnim(ui.AddMapAnim(gfx.NewAnim(0.0)), materialGib,
		ParticleBlastEmitter(3, 10),
		sx, sy, gibNum)

	PlaySound("death")
}

func (self *SdlEffects) Quit(message string) {
	MorePrompt()
	fmt.Print(message)
	Quit()
}

func (self *SdlEffects) MorePrompt() { MorePrompt() }

func (self *SdlEffects) Explode(center geom.Pt2I, power int, radius int) {
	sx, sy := CenterDrawPos(center)
	const gibNum = 20
	go ParticleAnim(ui.AddMapAnim(gfx.NewAnim(0.0)), materialSpark,
		smallBlastEmit, sx, sy, gibNum)

	// TODO: Explosion sound
}

func (self *SdlEffects) GetPlayerInput() (result (func() bool)) {
	ReleaseUISync()
	result = <-playerInputChan
	MarkMsgLinesSeen()
	GetUISync()
	return
}

func SendPlayerInput(command (func() bool)) bool {
	// Don't block, if the channel isn't expecting input, just move on and
	// return false.
	ok := playerInputChan <- command
	return ok
}

var playerInputChan = make(chan (func() bool))

func MorePrompt() {
	game.Msg("--more--\n")
	for key := GetKey(); !isActionKey(key) && key != ' '; key = GetKey() {
	}

	MarkMsgLinesSeen()
}

func SmartPlayerPickup(alwaysPickupFirst bool) entity.Id {
	itemIds := iterable.Data(game.TakeableItems(game.GetPos(game.PlayerId())))

	if len(itemIds) == 0 {
		game.Msg("Nothing to take here.\n")
		return entity.NilId
	}

	id := itemIds[0].(entity.Id)
	if len(itemIds) > 1 && !alwaysPickupFirst {
		id = EntityChoiceDialog("Pick up which item?", itemIds)
		if id == entity.NilId {
			game.Msg("Okay, then.\n")
			return entity.NilId
		}
	}
	SendPlayerInput(func() bool {
		if game.TakeItem(game.PlayerId(), id) {
			game.AutoEquip(game.PlayerId(), id)
		}
		return true
	})
	return id
}

func VisualScale() int { return config.TileScale }

func YesNoInput() bool {
	key := GetKey()
	return key == 'y' || key == 'Y'
}

const medkitCommandSlot = 0

func UsePower(idx int) {
	player := game.PlayerId()

	switch idx {
	case medkitCommandSlot:
		id, ok := game.FirstContentItem(player, game.IsMedKit)
		if !ok {
			game.Msg("You have no medkits.\n")
		} else {
			SendPlayerInput(func() bool {
				game.UseItem(player, id)
				return true
			})
		}
		return
	}
	game.Msg("Power %d is unbound.\n", idx)
}

// ShortNamePower returns the name of the power in the given slot in a form
// short enough to print to the on-screen command bar. Cooldowns and uses left
// can be included if applicable.
func ShortNamePower(idx int) string {
	player := game.PlayerId()
	switch idx {
	case medkitCommandSlot:
		medkitCount := game.ContentItemCount(player, game.IsMedKit)
		// TODO: Gray out the text when there are zero medkits.
		return fmt.Sprintf("medkit (%d)", medkitCount)
	}
	return ""
}
