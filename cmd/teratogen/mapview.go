package main

import (
	"exp/draw"
	"exp/iterable"
	"hyades/alg"
	"hyades/geom"
	"hyades/gfx"
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

func (self *MapView) InvTransform(area draw.Rectangle, screenX, screenY int) (worldX, worldY int) {
	worldX, worldY = Tile2WorldPos(GetWorld().GetPlayer().GetPos())
	worldX += screenX - area.Min.X - area.Dx()/2
	worldY += screenY - area.Min.Y - area.Dy()/2
	return
}

func (self *MapView) HandleMouseEvent(area draw.Rectangle, event draw.Mouse) bool {
	wx, wy := self.InvTransform(area, event.X, event.Y)
	go ParticleAnim(ui.AddMapAnim(gfx.NewAnim(0.0)), wx, wy,
		config.Scale, 1e8, float64(config.Scale)*20.0,
		gfx.White, gfx.Cyan, 6)

	return true
}

func (self *MapView) MouseExited(event draw.Mouse) {
}
