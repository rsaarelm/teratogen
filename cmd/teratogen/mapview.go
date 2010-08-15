package main

import (
	"container/vector"
	"exp/draw"
	"exp/iterable"
	"fmt"
	"hyades/alg"
	"hyades/dbg"
	"hyades/entity"
	"hyades/geom"
	"hyades/gfx"
	"hyades/keyboard"
	"hyades/num"
	"image"
	"rand"
	game "teratogen"
	"time"
)

var tileset = []int{
	game.TerrainIndeterminate: 255,
	game.TerrainFloor:         0,
	game.TerrainDoor:          3,
	game.TerrainStairDown:     4,
	game.TerrainCorridor:      13,
	game.TerrainWall:          16,
	game.TerrainDirtWall:      20,
	game.TerrainBrickWall:     24,
	game.TerrainRockWall:      28,
	game.TerrainBioWall:       32,
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
	lastArea      image.Rectangle

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
	return TileW*pos.X - TileW*pos.Y + xDrawOffset, TileH/2*pos.Y + TileH/2*pos.X + yDrawOffset
}

func CenterDrawPos(pos geom.Pt2I) (screenX, screenY int) {
	screenX, screenY = DrawPos(pos)
	screenX, screenY = screenX+TileW/2, screenY+TileH/2
	return
}

func Draw(g gfx.Graphics, spriteId string, x, y int) {
	sx, sy := DrawPos(geom.Pt2I{x, y})
	DrawSprite(g, spriteId, sx, sy)
}

func DrawShadow(g gfx.Graphics, darkness uint8, x, y int) {
	sx, sy := DrawPos(geom.Pt2I{x, y})
	g.FillRect(
		image.Rect(sx, sy, sx+TileW, sy+TileH),
		image.RGBAColor{0, 0, 0, darkness})
}

func DrawWorld(g gfx.Graphics) {
	drawTerrain(g)
	drawEntities(g)
	drawShadows(g)
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

	for _, sorted := range *seq {
		id := sorted.(entity.Id)
		pos := game.GetPos(id)
		seen := game.GetFov().Get(pos) == game.FovSeen
		mapped := seen || game.GetFov().Get(pos) == game.FovMapped
		esp := game.PlayerIsEsper() && game.CanEsperSense(id)
		// TODO: Draw static (item) entities from map memory.
		if esp || mapped {
			if esp || seen || !game.IsMobile(id) {
				var armor float64
				if crit := game.GetCreature(id); crit != nil {
					armor = crit.Armor
				}
				Draw(g, GearedIcon(game.GetIconId(id), armor), pos.X, pos.Y)
				continue
			}
		}
	}
}

func drawTerrain(g gfx.Graphics) {
	mapWidth, mapHeight := game.MapDims()
	for pt := range geom.PtIter(0, 0, mapWidth, mapHeight) {
		if game.GetFov().Get(pt) == game.FovUnknown {
			continue
		}
		idx := game.GetArea().GetTerrain(pt)
		tileIdx := tileset[idx]
		if isVisualWall(idx) && idx != game.TerrainDoor {
			mask := geom.HexNeighborMask(pt, func(p geom.Pt2I) bool { return isVisualWall(game.GetArea().GetTerrain(p)) })
			offset := geom.HexWallType(mask)

			tileIdx += offset
		}

		// XXX: Get indexable tilesets, so won't need these string kludges.
		tile := fmt.Sprintf("tiles:%d", tileIdx)
		Draw(g, tile, pt.X, pt.Y)
	}
}

func drawShadows(g gfx.Graphics) {
	mapWidth, mapHeight := game.MapDims()
	for pt := range geom.PtIter(0, 0, mapWidth, mapHeight) {
		if game.GetFov().Get(pt) == game.FovUnknown {
			continue
		}
		dist := geom.HexDist(game.GetPos(game.PlayerId()), pt)
		DrawShadow(g, uint8(num.Clamp(0, 255, float64(dist*12))), pt.X, pt.Y)
	}
}

func entityDrawOrder(i interface{}) int {
	id := i.(entity.Id)
	if game.IsDecal(id) {
		return -1
	}
	if game.IsCreature(id) {
		return 1
	}

	return 0
}

func entityEarlierInDrawOrder(i, j interface{}) bool {
	return entityDrawOrder(i) < entityDrawOrder(j)
}

func AnimTest() { go TestAnim2(ui.AddScreenAnim(gfx.NewAnim(0.0))) }

func (self *MapView) Draw(g gfx.Graphics, area image.Rectangle) {
	g.SetClip(area)
	defer g.ClearClip()

	elapsed := time.Nanoseconds() - self.timePoint
	self.timePoint += elapsed

	g2 := &gfx.TranslateGraphics{image.Pt(0, 0), g}
	x, y := CenterDrawPos(game.GetPos(game.PlayerId()))
	g2.Center(area, x, y)
	DrawWorld(g2)
	self.DrawAnims(g2, elapsed)
}

func (self *MapView) Children(area image.Rectangle) iterable.Iterable {
	return alg.EmptyIter()
}

func Tile2WorldPos(tilePos geom.Pt2I) geom.Pt2I {
	x, y := CenterDrawPos(tilePos)
	return geom.Pt2I{x, y}
}

func World2TilePos(worldPos geom.Pt2I) geom.Pt2I {
	return geom.Pt2I{
		(worldPos.X/2 + worldPos.Y) / TileW,
		(worldPos.Y - worldPos.X/2) / TileH}
}

func (self *MapView) InvTransform(area image.Rectangle, screenX, screenY int) (worldX, worldY int) {
	worldPos := Tile2WorldPos(game.GetPos(game.PlayerId()))
	worldPos.X += screenX - area.Min.X - area.Dx()/2
	worldPos.Y += screenY - area.Min.Y - area.Dy()/2
	return worldPos.X, worldPos.Y
}

func (self *MapView) onMouseButton(button int) {
	const mapEdit = false

	event := self.lastMouse
	area := self.lastArea
	wx, wy := self.InvTransform(area, event.X, event.Y)

	tilePos := World2TilePos(geom.Pt2I{wx, wy})
	vec := tilePos.Minus(game.GetPos(game.PlayerId()))
	switch button {
	case leftButton:
		if mapEdit {
			game.GetArea().SetTerrain(tilePos, game.TerrainDirtWall)
			return
		}

		// Move player in mouse direction.
		if !vec.Equals(geom.ZeroVec2I) {
			dir8 := geom.Vec2IToDir8(vec)
			SendPlayerInput(func() bool {
				game.SmartMovePlayer(dir8)
				return true
			})
		} else {
			// Clicking at player pos.

			// If there are stairs here, clicking on player goes down.
			if game.GetArea().GetTerrain(game.GetPos(game.PlayerId())) == game.TerrainStairDown {
				SendPlayerInput(func() bool {
					game.PlayerEnterStairs()
					return true
				})
			}

			// Pick up the first item so that mousing isn't interrupted by a
			// currently keyboard-only dialog if there are many items.

			// TODO: Support choosing which item to pick up when using mouse.
			SmartPlayerPickup(true)
		}
	case rightButton:
		if mapEdit {
			game.GetArea().SetTerrain(tilePos, game.TerrainFloor)
			return
		}

		if !vec.Equals(geom.ZeroVec2I) {
			// TODO: Shoot enemies if they are on a firing line.
		}
	}
}

func playerShootDir(dir6 int) {
	SendPlayerInput(func() bool {
		id := game.PlayerId()
		dist := 1
		if weapon := game.GetCreature(id).Weapon2(); weapon != nil {
			dist = weapon.Range
		}
		targ := game.GetPos(id).Plus(geom.Dir6ToVec(dir6).Scale(dist))
		return game.Shoot(id, targ)
	})
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

func (self *MapView) HandleMouseEvent(area image.Rectangle, event draw.Mouse) bool {
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
		1, 1e8, float64(config.Scale)*20.0,
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
		SendPlayerInput(func() bool { return true })
	case 'Q':
		game.Msg("Really quit? Your game will be lost. [y/n]\n")
		if YesNoInput() {
			Quit()
		} else {
			game.Msg("Okay, then.\n")
		}
	case 'w', keyboard.K_KP8, keyboard.K_HOME, keyboard.K_UP:
		SendPlayerInput(func() bool {
			game.SmartMovePlayer(0)
			return true
		})
	case 'e', keyboard.K_PAGEUP, keyboard.K_KP9:
		SendPlayerInput(func() bool {
			game.SmartMovePlayer(1)
			return true
		})
	case keyboard.K_RIGHT, keyboard.K_KP6:
		SendPlayerInput(func() bool {
			game.SmartMovePlayer(2)
			return true
		})
	case 'd', keyboard.K_PAGEDOWN, keyboard.K_KP3:
		SendPlayerInput(func() bool {
			game.SmartMovePlayer(3)
			return true
		})
	case 's', keyboard.K_KP2, keyboard.K_END, keyboard.K_DOWN:
		SendPlayerInput(func() bool {
			game.SmartMovePlayer(4)
			return true
		})
	case 'a', keyboard.K_DELETE, keyboard.K_KP1:
		SendPlayerInput(func() bool {
			game.SmartMovePlayer(5)
			return true
		})
	case keyboard.K_LEFT, keyboard.K_KP4:
		SendPlayerInput(func() bool {
			game.SmartMovePlayer(6)
			return true
		})
	case keyboard.K_RETURN, keyboard.K_KP5:
		SmartLocalPlayerAction()
	case 'q', keyboard.K_INSERT, keyboard.K_KP7:
		SendPlayerInput(func() bool {
			game.SmartMovePlayer(7)
			return true
		})
	case 'u':
		playerShootDir(5)
	case 'i':
		playerShootDir(0)
	case 'o':
		playerShootDir(1)
	case 'l':
		playerShootDir(2)
	case 'k':
		playerShootDir(3)
	case 'j':
		playerShootDir(4)
	case 'A':
		ApplyItemMenu()
	case ',':
		SmartPlayerPickup(false)
	case 't':
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
	case 'E':
		EquipMenu()
	case 'D':
		// Drop item.
		if game.HasContents(game.PlayerId()) {
			id := EntityChoiceDialog(
				"Drop which item?", iterable.Data(game.Contents(game.PlayerId())))
			if id != entity.NilId {
				SendPlayerInput(func() bool {
					game.DropItem(game.PlayerId(), id)
					return true
				})
			} else {
				game.Msg("Okay, then.\n")
			}
		} else {
			game.Msg("Nothing to drop.\n")
		}
	case '>':
		SendPlayerInput(func() bool {
			game.PlayerEnterStairs()
			return true
		})
	case 'S':
		err := game.SaveGame(SaveFileName(), UseGzipSaves)
		dbg.AssertNoError(err)
		game.Msg("Game saved.\n")
		fmt.Println("Be seeing you.")
		Quit()
	case keyboard.K_F12:
		game.Msg("Saving screenshot.\n")
		SaveScreenshot()
	}
}

func (self *MapView) HandleKey(key int) { go self.AsyncHandleKey(key) }

func (self *MapView) MouseExited(event draw.Mouse) {
	for i := 0; i < 3; i++ {
		self.releaseMouse(i)
	}
}

func isVisualWall(terrain game.TerrainType) bool {
	return terrain >= game.TerrainDoor
}

// InCellJitter return a vector that point's to a normal-distributed position
// within a game tile centered on tile center.
func InCellJitter() geom.Vec2I {
	x, y := rand.NormFloat64()*float64(TileW)/4, rand.NormFloat64()*float64(TileH)/4
	x, y = num.Clamp(
		-float64(TileW)/2, float64(TileW)/2, x),
		num.Clamp(-float64(TileH)/2, float64(TileH)/2, y)
	return geom.Vec2I{int(x), int(y)}
}

// GearedIcon may change the icon of a character based on it's gear.
func GearedIcon(icon string, armorValue float64) string {
	armorLevel := int(armorValue * game.ArmorScale)
	if icon == "chars:16" {
		if armorLevel > game.RiotArmorLevel {
			// Hard suit
			return "chars:19"
		} else if armorLevel > game.VestArmorLevel {
			// Riot armor
			return "chars:18"
		} else if armorLevel > 0 {
			// Tactical vest
			return "chars:17"
		}
	}

	return icon
}

// SmartLocalPlayerAction looks for an obvious action to do at the local
// position. If there are items to pick up, it offers to pick them up. If
// there are stairs, it goes down stairs. Return whether an action was
// performed.
func SmartLocalPlayerAction() {
	if game.NumPlayerTakeableItems() > 0 {
		item := SmartPlayerPickup(false)
		if item != entity.NilId {
			return
		}
	}
	if game.PlayerAtStairs() {
		SendPlayerInput(func() bool {
			game.PlayerEnterStairs()
			return true
		})
		return
	}
	// Default is to idle.
	SendPlayerInput(func() bool { return true })
}
