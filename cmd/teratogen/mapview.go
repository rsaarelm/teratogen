package main

import (
	"exp/draw"
	"exp/iterable"
	"hyades/alg"
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

func AnimTest() { go TestAnim(ui.context, ui.AddScreenAnim(gfx.NewAnim(0.0))) }

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
