package main

import (
	"container/vector"
	"exp/draw"
	"exp/iterable"
	"hyades/alg"
	"hyades/gui"
	"time"
)

const TileW = 16
const TileH = 16

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

func animSort(i, j interface{}) bool { return i.(*Anim).Z < j.(*Anim).Z }

func AnimTest() { go TestAnim(ui.context, ui.AddAnim(NewAnim(0.0))) }

func (self *MapView) AddAnim(anim *Anim) *Anim {
	self.anims.Push(anim)
	return anim
}

// TODO: Pass draw offset to anims.

func (self *MapView) DrawAnims(timeElapsedNs int64) {
	alg.PredicateSort(animSort, self.anims)
	for i := 0; i < self.anims.Len(); i++ {
		anim := self.anims.At(i).(*Anim)
		if anim.Closed() {
			self.anims.Delete(i)
			i--
			continue
		}
		// Tell the anim it can draw itself.
		anim.UpdateChan <- timeElapsedNs
		// Wait for the anim to call back that it's completed drawing itself.
		<-anim.UpdateChan
	}
}

func (self *MapView) Draw(g gui.Graphics, area draw.Rectangle) {
	elapsed := time.Nanoseconds() - self.timePoint
	self.timePoint += elapsed

	// TODO: Local, custom world draw.
	// TODO: Adapt world draw to area.
	world := GetWorld()
	world.Draw()
	self.DrawAnims(elapsed)
}

func (self *MapView) Children(area draw.Rectangle) iterable.Iterable {
	return alg.EmptyIter()
}
