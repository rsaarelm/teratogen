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
	"os"
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
		esp := game.PlayerIsEsper() && game.CanEsperSense(id)
		// TODO: Draw static (item) entities from map memory.
		if esp || mapped {
			if esp || seen || !game.IsMobile(id) {
				armorId, _ := game.GetEquipment(id, game.ArmorEquipSlot)
				Draw(g, GearedIcon(game.GetIconId(id), armorId), pos.X, pos.Y)
				continue
			}
		}
	}
}

func drawTerrain(g gfx.Graphics) {
	mapWidth, mapHeight := game.MapDims()
	for pt := range geom.PtIter(0, 0, mapWidth, mapHeight) {
		if game.GetLos().Get(pt) == game.LosUnknown {
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

func (self *MapView) Draw(g gfx.Graphics, area draw.Rectangle) {
	g.SetClip(area)
	defer g.ClearClip()

	elapsed := time.Nanoseconds() - self.timePoint
	self.timePoint += elapsed

	g2 := &gfx.TranslateGraphics{draw.Pt(0, 0), g}
	x, y := CenterDrawPos(game.GetPos(game.PlayerId()))
	g2.Center(area, x, y)
	DrawWorld(g2)
	self.DrawAnims(g2, elapsed)
}

func (self *MapView) Children(area draw.Rectangle) iterable.Iterable {
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

func (self *MapView) InvTransform(area draw.Rectangle, screenX, screenY int) (worldX, worldY int) {
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
			// If shift is pressed, right-click always shoots.
			if self.shiftKeyState > 0 {
				SendPlayerInput(func() bool {
					game.Shoot(game.PlayerId(), tilePos)
					return true
				})
				return
			}

			// If there's an enemy, right-click shoots at it.
			for _ = range game.EnemiesAt(game.PlayerId(), tilePos).Iter() {
				SendPlayerInput(func() bool {
					game.Shoot(game.PlayerId(), tilePos)
					return true
				})
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
		SendPlayerInput(func() bool { return true })
	case 'q':
		Quit()
	case 'i', keyboard.K_KP8, keyboard.K_HOME, keyboard.K_UP:
		SendPlayerInput(func() bool {
			game.SmartMovePlayer(0)
			return true
		})
	case 'o', keyboard.K_PAGEUP, keyboard.K_KP9:
		SendPlayerInput(func() bool {
			game.SmartMovePlayer(1)
			return true
		})
	case keyboard.K_RIGHT, keyboard.K_KP6:
		SendPlayerInput(func() bool {
			game.SmartMovePlayer(2)
			return true
		})
	case 'l', keyboard.K_PAGEDOWN, keyboard.K_KP3:
		SendPlayerInput(func() bool {
			game.SmartMovePlayer(3)
			return true
		})
	case 'k', keyboard.K_KP2, keyboard.K_END, keyboard.K_DOWN:
		SendPlayerInput(func() bool {
			game.SmartMovePlayer(4)
			return true
		})
	case 'j', keyboard.K_DELETE, keyboard.K_KP1:
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
	case 'u', keyboard.K_INSERT, keyboard.K_KP7:
		SendPlayerInput(func() bool {
			game.SmartMovePlayer(7)
			return true
		})
	case 'a':
		if ApplyItemMenu() {
			SendPlayerInput(func() bool { return true })
		}
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
	case 'e':
		EquipMenu()
	case 'f':
		if game.GunEquipped(game.PlayerId()) {
			targetId := game.ClosestCreatureSeenBy(game.PlayerId())
			if targetId != entity.NilId {
				SendPlayerInput(func() bool { return game.Shoot(game.PlayerId(), game.GetPos(targetId)) })
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
		loadFile.Close()
		game.Msg("Game loaded.\n")
		ReleaseUISync()
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
func GearedIcon(icon string, armor entity.Id) string {
	if icon == "chars:16" {
		// Player icon
		switch game.GetName(armor) {
		case "kevlar armor":
			return "chars:17"
		case "riot armor":
			return "chars:18"
		case "hard suit":
			return "chars:19"
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
	game.Msg("Nothing to do here.\n")
}
