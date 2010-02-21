package main

import (
	"container/vector"
	"exp/draw"
	"exp/iterable"
	"hyades/alg"
	"hyades/dbg"
	"hyades/entity"
	"hyades/geom"
	"hyades/gfx"
	"hyades/keyboard"
	"hyades/num"
	"os"
	"time"
)

const mouseRepeatInterval = 150e6

const (
	leftButton   = 0
	middleButton = 1
	rightButton  = 2
)

type MapView struct {
	gfx.Anims
	timePoint int64
	// Keeps track of the states of the three mouse buttons. -1 when the button
	// is up, the time to spawn the next mouse-button-down event otherwise.
	// XXX: Whether a mouse button should spawn periodic events when pressed is
	// context-dependent, should have a more general scheduling mechanism for
	// this.
	mouseDownTime [3]int64
	lastMouse     draw.Mouse
	lastArea      draw.Rectangle

	// XXX: Hacky way to remember if shift is pressed.
	shiftKeyState int
}

func NewMapView() (result *MapView) {
	result = new(MapView)
	result.InitAnims()
	result.timePoint = time.Nanoseconds()
	result.mouseDownTime = [3]int64{-1, -1, -1}
	return
}

func DrawPos(pos geom.Pt2I) (screenX, screenY int) {
	return TileW*pos.X + xDrawOffset, TileH*pos.Y + yDrawOffset
}

func CenterDrawPos(pos geom.Pt2I) (screenX, screenY int) {
	return TileW*pos.X + xDrawOffset + TileW/2, TileH*pos.Y + yDrawOffset + TileH/2
}

func Draw(g gfx.Graphics, spriteId string, x, y int) {
	sx, sy := DrawPos(geom.Pt2I{x, y})
	DrawSprite(g, spriteId, sx, sy)
}

func DrawWorld(g gfx.Graphics) {
	drawTerrain(g)
	drawEntities(g)
}

func drawEntities(g gfx.Graphics) {
	// Make a vector of the entities sorted in draw order.
	seq := new(vector.Vector)
	for o := range Entities().Iter() {
		id := o.(entity.Id)
		ent := GetBlob(id)
		if ent == nil {
			continue
		}
		if GetParent(ent.GetGuid()) != entity.NilId {
			// Skip entities inside something.
			continue
		}
		seq.Push(ent)
	}
	alg.PredicateSort(entityEarlierInDrawOrder, seq)

	for sorted := range seq.Iter() {
		e := sorted.(*Blob)
		pos := GetPos(e.GetGuid())
		seen := GetLos().Get(pos) == LosSeen
		mapped := seen || GetLos().Get(pos) == LosMapped
		// TODO: Draw static (item) entities from map memory.
		if mapped {
			if seen || !IsMobile(e.GetGuid()) {
				Draw(g, e.IconId, pos.X, pos.Y)
			}
		}
	}
}

func entityEarlierInDrawOrder(i, j interface{}) bool {
	return i.(*Blob).GetClass() < j.(*Blob).GetClass()
}

func AnimTest() { go TestAnim2(ui.AddScreenAnim(gfx.NewAnim(0.0))) }

func (self *MapView) Draw(g gfx.Graphics, area draw.Rectangle) {
	g.SetClip(area)
	defer g.ClearClip()

	elapsed := time.Nanoseconds() - self.timePoint
	self.timePoint += elapsed

	g2 := &gfx.TranslateGraphics{draw.Pt(0, 0), g}
	g2.Center(area, GetPos(PlayerId()).X*TileW+TileW/2,
		GetPos(PlayerId()).Y*TileH+TileH/2)
	DrawWorld(g2)
	self.DrawAnims(g2, elapsed)
}

func (self *MapView) Children(area draw.Rectangle) iterable.Iterable {
	return alg.EmptyIter()
}

func Tile2WorldPos(tilePos geom.Pt2I) (worldX, worldY int) {
	return tilePos.X*TileW + TileW/2, tilePos.Y*TileH + TileH/2
}

func World2TilePos(worldX, worldY int) geom.Pt2I {
	return geom.Pt2I{worldX / TileW, worldY / TileH}
}

func (self *MapView) InvTransform(area draw.Rectangle, screenX, screenY int) (worldX, worldY int) {
	worldX, worldY = Tile2WorldPos(GetPos(PlayerId()))
	worldX += screenX - area.Min.X - area.Dx()/2
	worldY += screenY - area.Min.Y - area.Dy()/2
	return
}

func (self *MapView) onMouseButton(button int) {
	event := self.lastMouse
	area := self.lastArea
	wx, wy := self.InvTransform(area, event.X, event.Y)

	tilePos := World2TilePos(wx, wy)
	vec := tilePos.Minus(GetPos(PlayerId()))
	switch button {
	case leftButton:
		// Move player in mouse direction.
		if !vec.Equals(geom.ZeroVec2I) {
			dir8 := geom.Vec2IToDir8(vec)
			SendPlayerInput(func() { SmartMovePlayer(dir8) })
		} else {
			// Clicking at player pos.

			// If there are stairs here, clicking on player goes down.
			if GetArea().GetTerrain(GetPos(PlayerId())) == TerrainStairDown {
				SendPlayerInput(func() { PlayerEnterStairs() })
			}

			// Pick up the first item so that mousing isn't interrupted by a
			// currently keyboard-only dialog if there are many items.

			// TODO: Support choosing which item to pick up when using mouse.
			if SmartPlayerPickup(true) != nil {
				SendPlayerInput(func() {})
			}
		}
	case rightButton:
		if !vec.Equals(geom.ZeroVec2I) {
			// If shift is pressed, right-click always shoots.
			if self.shiftKeyState > 0 {
				SendPlayerInput(func() { Shoot(PlayerId(), tilePos) })
				return
			}

			// If there's an enemy, right-click shoots at it.
			for _ = range EnemiesAt(PlayerId(), tilePos).Iter() {
				SendPlayerInput(func() { Shoot(PlayerId(), tilePos) })
				return
			}
		}
	}
}

func (self *MapView) isMousePressed(button int) bool {
	return self.mouseDownTime[button] >= 0
}

func (self *MapView) pressMouse(button int) {
	if !self.isMousePressed(button) {
		self.onMouseButton(button)
		self.mouseDownTime[button] = time.Nanoseconds() + mouseRepeatInterval
	}
}

func (self *MapView) releaseMouse(button int) { self.mouseDownTime[button] = -1 }

func (self *MapView) HandleMouseEvent(area draw.Rectangle, event draw.Mouse) bool {
	wx, wy := self.InvTransform(area, event.X, event.Y)

	self.lastMouse = event
	self.lastArea = area

	for i := 0; i < 3; i++ {
		if event.Buttons&(1<<byte(i)) != 0 {
			self.pressMouse(i)
		} else {
			self.releaseMouse(i)
		}
	}

	go ParticleAnim(ui.AddMapAnim(gfx.NewAnim(0.0)), wx, wy,
		config.Scale, 1e8, float64(config.Scale)*20.0,
		gfx.White, gfx.Cyan, 6)

	return true
}

func (self *MapView) HandleTickEvent(elapsedNs int64) {
	t := time.Nanoseconds()

	for i := 0; i < 3; i++ {
		// Mouse repeater logic.
		if self.isMousePressed(i) && self.mouseDownTime[i] <= t {
			self.mouseDownTime[i] = -1
			self.pressMouse(i)
		}
	}
}

func (self *MapView) AsyncHandleKey(key int) {
	key = keymap.Map(key)

	// Key release events have negated keycodes, so taking abs value gets
	// the event key regardless of whether it was pressed or released.
	if abskey := num.Iabs(key); abskey == keyboard.K_RSHIFT || abskey == keyboard.K_LSHIFT {
		self.shiftKeyState += key
		if self.shiftKeyState < 0 {
			// Hack to fix starting with the shift key pressed and being in an
			// incosistent state. Zeroing out after going negative should help
			// normalize the state.
			self.shiftKeyState = 0
		}
	}

	switch key {
	case '.':
		// Idle.
		SendPlayerInput(func() {})
	case 'q':
		Quit()
	case 'k', keyboard.K_UP, keyboard.K_KP8:
		SendPlayerInput(func() { SmartMovePlayer(0) })
	case 'u', keyboard.K_PAGEUP, keyboard.K_KP9:
		SendPlayerInput(func() { SmartMovePlayer(1) })
	case 'l', keyboard.K_RIGHT, keyboard.K_KP6:
		SendPlayerInput(func() { SmartMovePlayer(2) })
	case 'n', keyboard.K_PAGEDOWN, keyboard.K_KP3:
		SendPlayerInput(func() { SmartMovePlayer(3) })
	case 'j', keyboard.K_DOWN, keyboard.K_KP2:
		SendPlayerInput(func() { SmartMovePlayer(4) })
	case 'b', keyboard.K_END, keyboard.K_KP1:
		SendPlayerInput(func() { SmartMovePlayer(5) })
	case 'h', keyboard.K_LEFT, keyboard.K_KP4:
		SendPlayerInput(func() { SmartMovePlayer(6) })
	case 'y', keyboard.K_HOME, keyboard.K_KP7:
		SendPlayerInput(func() { SmartMovePlayer(7) })
	case 'a':
		if ApplyItemMenu() {
			SendPlayerInput(func() {})
		}
	case ',':
		if SmartPlayerPickup(false) != nil {
			SendPlayerInput(func() {})
		}
	case 'i':
		// Show inventory.
		Msg("Carried:")
		first := true
		for o := range Contents(GetBlob(PlayerId()).GetGuid()).Iter() {
			item := GetBlob(o.(entity.Id))
			if first {
				first = false
				Msg(" %v", GetName(item.GetGuid()))
			} else {
				Msg(", %v", GetName(item.GetGuid()))
			}
		}
		if first {
			Msg(" nothing.\n")
		} else {
			Msg(".\n")
		}
	case 'e':
		EquipMenu()
	case 'f':
		if GunEquipped(GetBlob(PlayerId())) {
			targetId := ClosestCreatureSeenBy(PlayerId())
			if targetId != entity.NilId {
				SendPlayerInput(func() { Shoot(PlayerId(), GetPos(targetId)) })
			} else {
				Msg("You see nothing to shoot.\n")
			}
		} else {
			Msg("You don't have a gun to fire.\n")
		}
	case 'd':
		// Drop item.
		if HasContents(PlayerId()) {
			item, ok := ObjectChoiceDialog(
				"Drop which item?", iterable.Data(iterable.Map(Contents(PlayerId()), id2Blob)))
			if ok {
				SendPlayerInput(func() {
					item := item.(*Blob)
					SetParent(item.GetGuid(), entity.NilId)
					Msg("Dropped %v.\n", GetName(item.GetGuid()))
				})
			} else {
				Msg("Okay, then.\n")
			}
		} else {
			Msg("Nothing to drop.\n")
		}
	case '>':
		SendPlayerInput(func() { PlayerEnterStairs() })
	case 'S':
		saveFile, err := os.Open("/tmp/saved.gam", os.O_WRONLY|os.O_CREAT, 0666)
		dbg.AssertNoError(err)
		GetContext().Serialize(saveFile)
		saveFile.Close()
		Msg("Game saved.\n")
	case 'L':
		GetUISync()
		loadFile, err := os.Open("/tmp/saved.gam", os.O_RDONLY, 0666)
		if err != nil {
			Msg("Error loading game: " + err.String())
			break
		}
		GetContext().Deserialize(loadFile)
		Msg("Game loaded.\n")
		ReleaseUISync()
	}
}

func (self *MapView) HandleKey(key int) { go self.AsyncHandleKey(key) }

func (self *MapView) MouseExited(event draw.Mouse) {
	for i := 0; i < 3; i++ {
		self.releaseMouse(i)
	}
}
