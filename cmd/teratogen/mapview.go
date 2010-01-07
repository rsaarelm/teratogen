package main

import (
	"exp/draw"
	"exp/iterable"
	"hyades/alg"
	"hyades/dbg"
	"hyades/geom"
	"hyades/gfx"
	"hyades/keyboard"
	"os"
	"time"
)

type MapView struct {
	gfx.Anims
	timePoint int64
}

func NewMapView() (result *MapView) {
	result = new(MapView)
	result.InitAnims()
	result.timePoint = time.Nanoseconds()
	return
}

func AnimTest() { go TestAnim2(ui.AddScreenAnim(gfx.NewAnim(0.0))) }

func (self *MapView) Draw(g gfx.Graphics, area draw.Rectangle) {
	g.SetClip(area)
	defer g.ClearClip()

	elapsed := time.Nanoseconds() - self.timePoint
	self.timePoint += elapsed

	world := GetWorld()
	g2 := &gfx.TranslateGraphics{draw.Pt(0, 0), g}
	g2.Center(area, world.GetPlayer().GetPos().X*TileW+TileW/2,
		world.GetPlayer().GetPos().Y*TileH+TileH/2)
	world.Draw(g2)
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
	worldX, worldY = Tile2WorldPos(GetWorld().GetPlayer().GetPos())
	worldX += screenX - area.Min.X - area.Dx()/2
	worldY += screenY - area.Min.Y - area.Dy()/2
	return
}

func (self *MapView) HandleMouseEvent(area draw.Rectangle, event draw.Mouse) bool {
	wx, wy := self.InvTransform(area, event.X, event.Y)

	tilePos := World2TilePos(wx, wy)
	vec := tilePos.Minus(GetWorld().GetPlayer().GetPos())
	if !vec.Equals(geom.ZeroVec2I) && (event.Buttons&1 != 0) {
		// TODO: Better handling for keeping the button pressed.
		dir8 := geom.Vec2IToDir8(vec)
		SendPlayerInput(func() { SmartMovePlayer(dir8) })
	}

	go ParticleAnim(ui.AddMapAnim(gfx.NewAnim(0.0)), wx, wy,
		config.Scale, 1e8, float64(config.Scale)*20.0,
		gfx.White, gfx.Cyan, 6)

	return true
}

func (self *MapView) HandleKey(key int) {
	key = keymap.Map(key)

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
		if SmartPlayerPickup() != nil {
			SendPlayerInput(func() {})
		}
	case 'i':
		// Show inventory.
		Msg("Carried:")
		first := true
		item := world.GetPlayer().GetChild()
		for item != nil {
			if first {
				first = false
				Msg(" %v", item.Name)
			} else {
				Msg(", %v", item.Name)
			}
			item = item.GetSibling()
		}
		if first {
			Msg(" nothing.\n")
		} else {
			Msg(".\n")
		}
	case 'e':
		EquipMenu()
	case 'f':
		player := world.GetPlayer()
		if GunEquipped(player) {
			target := ClosestCreatureSeenBy(player)
			if target != nil {
				SendPlayerInput(func() { Shoot(player, target.GetPos()) })
			} else {
				Msg("You see nothing to shoot.\n")
			}
		} else {
			Msg("You don't have a gun to fire.\n")
		}
	case 'd':
		// Drop item.
		player := world.GetPlayer()
		if player.HasContents() {
			item, ok := ObjectChoiceDialog(
				"Drop which item?", iterable.Data(player.Contents()))
			if ok {
				SendPlayerInput(func() {
					item := item.(*Entity)
					item.RemoveSelf()
					Msg("Dropped %v.\n", item.GetName())
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
		world.Serialize(saveFile)
		saveFile.Close()
		Msg("Game saved.\n")
	case 'L':
		loadFile, err := os.Open("/tmp/saved.gam", os.O_RDONLY, 0666)
		if err != nil {
			Msg("Error loading game: " + err.String())
			break
		}
		world = new(World)
		SetWorld(world)
		world.Deserialize(loadFile)
		Msg("Game loaded.\n")
	}

}

func (self *MapView) MouseExited(event draw.Mouse) {
}
