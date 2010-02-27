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
	game "teratogen"
	"time"
)

var tileset1 = []string{
	game.TerrainIndeterminate: "tiles:255",
	game.TerrainWall: "tiles:8",
	game.TerrainWallFront: "tiles:7",
	game.TerrainFloor: "tiles:0",
	game.TerrainDoor: "tiles:3",
	game.TerrainStairDown: "tiles:4",
	game.TerrainDirt: "tiles:6",
	game.TerrainDirtFront: "tiles:5",
}


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
	for o := range game.Entities().Iter() {
		id := o.(entity.Id)
		if id == entity.NilId {
			continue
		}
		if !game.HasPosComp(id) {
			continue
		}
		if game.GetParent(id) != entity.NilId {
			// Skip entities inside something.
			continue
		}
		seq.Push(id)
	}
	alg.PredicateSort(entityEarlierInDrawOrder, seq)

	for sorted := range seq.Iter() {
		id := sorted.(entity.Id)
		pos := game.GetPos(id)
		seen := game.GetLos().Get(pos) == game.LosSeen
		mapped := seen || game.GetLos().Get(pos) == game.LosMapped
		// TODO: Draw static (item) entities from map memory.
		if mapped {
			if seen || !game.IsMobile(id) {
				Draw(g, game.GetIconId(id), pos.X, pos.Y)
			}
		}
	}
}

func entityEarlierInDrawOrder(i, j interface{}) bool {
	return !game.IsCreature(i.(entity.Id)) && game.IsCreature(j.(entity.Id))
}

func AnimTest() { go TestAnim2(ui.AddScreenAnim(gfx.NewAnim(0.0))) }

func (self *MapView) Draw(g gfx.Graphics, area draw.Rectangle) {
	g.SetClip(area)
	defer g.ClearClip()

	elapsed := time.Nanoseconds() - self.timePoint
	self.timePoint += elapsed

	g2 := &gfx.TranslateGraphics{draw.Pt(0, 0), g}
	g2.Center(area, game.GetPos(game.PlayerId()).X*TileW+TileW/2,
		game.GetPos(game.PlayerId()).Y*TileH+TileH/2)
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
	worldX, worldY = Tile2WorldPos(game.GetPos(game.PlayerId()))
	worldX += screenX - area.Min.X - area.Dx()/2
	worldY += screenY - area.Min.Y - area.Dy()/2
	return
}

func (self *MapView) onMouseButton(button int) {
	event := self.lastMouse
	area := self.lastArea
	wx, wy := self.InvTransform(area, event.X, event.Y)

	tilePos := World2TilePos(wx, wy)
	vec := tilePos.Minus(game.GetPos(game.PlayerId()))
	switch button {
	case leftButton:
		// Move player in mouse direction.
		if !vec.Equals(geom.ZeroVec2I) {
			dir8 := geom.Vec2IToDir8(vec)
			game.SendPlayerInput(func() { game.SmartMovePlayer(dir8) })
		} else {
			// Clicking at player pos.

			// If there are stairs here, clicking on player goes down.
			if game.GetArea().GetTerrain(game.GetPos(game.PlayerId())) == game.TerrainStairDown {
				game.SendPlayerInput(func() { game.PlayerEnterStairs() })
			}

			// Pick up the first item so that mousing isn't interrupted by a
			// currently keyboard-only dialog if there are many items.

			// TODO: Support choosing which item to pick up when using mouse.
			SmartPlayerPickup(true)
		}
	case rightButton:
		if !vec.Equals(geom.ZeroVec2I) {
			// If shift is pressed, right-click always shoots.
			if self.shiftKeyState > 0 {
				game.SendPlayerInput(func() { game.Shoot(game.PlayerId(), tilePos) })
				return
			}

			// If there's an enemy, right-click shoots at it.
			for _ = range game.EnemiesAt(game.PlayerId(), tilePos).Iter() {
				game.SendPlayerInput(func() { game.Shoot(game.PlayerId(), tilePos) })
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
		game.SendPlayerInput(func() {})
	case 'q':
		Quit()
	case 'k', keyboard.K_UP, keyboard.K_KP8:
		game.SendPlayerInput(func() { game.SmartMovePlayer(0) })
	case 'u', keyboard.K_PAGEUP, keyboard.K_KP9:
		game.SendPlayerInput(func() { game.SmartMovePlayer(1) })
	case 'l', keyboard.K_RIGHT, keyboard.K_KP6:
		game.SendPlayerInput(func() { game.SmartMovePlayer(2) })
	case 'n', keyboard.K_PAGEDOWN, keyboard.K_KP3:
		game.SendPlayerInput(func() { game.SmartMovePlayer(3) })
	case 'j', keyboard.K_DOWN, keyboard.K_KP2:
		game.SendPlayerInput(func() { game.SmartMovePlayer(4) })
	case 'b', keyboard.K_END, keyboard.K_KP1:
		game.SendPlayerInput(func() { game.SmartMovePlayer(5) })
	case 'h', keyboard.K_LEFT, keyboard.K_KP4:
		game.SendPlayerInput(func() { game.SmartMovePlayer(6) })
	case 'y', keyboard.K_HOME, keyboard.K_KP7:
		game.SendPlayerInput(func() { game.SmartMovePlayer(7) })
	case 'a':
		if ApplyItemMenu() {
			game.SendPlayerInput(func() {})
		}
	case ',':
		SmartPlayerPickup(false)
	case 'i':
		// Show inventory.
		game.Msg("Carried:")
		first := true
		for o := range game.Contents(game.PlayerId()).Iter() {
			id := o.(entity.Id)
			if first {
				first = false
				game.Msg(" %v", game.GetName(id))
			} else {
				game.Msg(", %v", game.GetName(id))
			}
		}
		if first {
			game.Msg(" nothing.\n")
		} else {
			game.Msg(".\n")
		}
	case 'e':
		EquipMenu()
	case 'f':
		if game.GunEquipped(game.PlayerId()) {
			targetId := game.ClosestCreatureSeenBy(game.PlayerId())
			if targetId != entity.NilId {
				game.SendPlayerInput(func() { game.Shoot(game.PlayerId(), game.GetPos(targetId)) })
			} else {
				game.Msg("You see nothing to shoot.\n")
			}
		} else {
			game.Msg("You don't have a gun to fire.\n")
		}
	case 'd':
		// Drop item.
		if game.HasContents(game.PlayerId()) {
			id := EntityChoiceDialog(
				"Drop which item?", iterable.Data(game.Contents(game.PlayerId())))
			if id != entity.NilId {
				game.SendPlayerInput(func() { game.DropItem(game.PlayerId(), id) })
			} else {
				game.Msg("Okay, then.\n")
			}
		} else {
			game.Msg("Nothing to drop.\n")
		}
	case '>':
		game.SendPlayerInput(func() { game.PlayerEnterStairs() })
	case 'S':
		saveFile, err := os.Open("/tmp/saved.gam", os.O_WRONLY|os.O_CREAT, 0666)
		dbg.AssertNoError(err)
		game.GetContext().Serialize(saveFile)
		saveFile.Close()
		game.Msg("Game saved.\n")
	case 'L':
		GetUISync()
		loadFile, err := os.Open("/tmp/saved.gam", os.O_RDONLY, 0666)
		if err != nil {
			game.Msg("Error loading game: " + err.String())
			break
		}
		game.GetContext().Deserialize(loadFile)
		game.Msg("Game loaded.\n")
		ReleaseUISync()
	}
}

func (self *MapView) HandleKey(key int) { go self.AsyncHandleKey(key) }

func (self *MapView) MouseExited(event draw.Mouse) {
	for i := 0; i < 3; i++ {
		self.releaseMouse(i)
	}
}

func drawTerrain(g gfx.Graphics) {
	mapWidth, mapHeight := game.MapDims()
	for pt := range geom.PtIter(0, 0, mapWidth, mapHeight) {
		if game.GetLos().Get(pt) == game.LosUnknown {
			continue
		}
		idx := game.GetArea().GetTerrain(pt)
		front := game.GetArea().GetTerrain(pt.Plus(geom.Vec2I{0, 1}))
		// XXX: Hack to get the front tile visuals
		if idx == game.TerrainWall && front != game.TerrainWall && front != game.TerrainDoor {
			idx = game.TerrainWallFront
		}
		if idx == game.TerrainDirt && front != game.TerrainDirt && front != game.TerrainDoor {
			idx = game.TerrainDirtFront
		}
		Draw(g, tileset1[idx], pt.X, pt.Y)
	}
}
