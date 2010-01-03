package main

import (
	"container/vector"
	"exp/draw"
	"exp/iterable"
	"hyades/alg"
	"hyades/gfx"
	"time"
)

type MapView struct {
	timePoint int64
	anims     *vector.Vector
}

func NewMapView() (result *MapView) {
	result = new(MapView)
	result.anims = new(vector.Vector)
	result.timePoint = time.Nanoseconds()
	return
}

func animSort(i, j interface{}) bool { return i.(*gfx.Anim).Z < j.(*gfx.Anim).Z }

func AnimTest() { go TestAnim(ui.context, ui.AddAnim(gfx.NewAnim(0.0))) }

func (self *MapView) AddAnim(anim *gfx.Anim) *gfx.Anim {
	self.anims.Push(anim)
	return anim
}

// TODO: Pass draw offset to anims.

func (self *MapView) DrawAnims(g gfx.Graphics, timeElapsedNs int64) {
	alg.PredicateSort(animSort, self.anims)
	for i := 0; i < self.anims.Len(); i++ {
		anim := self.anims.At(i).(*gfx.Anim)
		if anim.Closed() {
			self.anims.Delete(i)
			i--
			continue
		}
		anim.Update(g, timeElapsedNs)
	}
}

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
